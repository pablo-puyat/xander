import json
import os
import re

import google.generativeai as genai
from github import Auth, Github


def parse_diff_to_changed_lines(patch):
    """
    Parses a git diff patch to identify which lines in the new file were added or modified.
    Returns a set of line numbers (integers) in the new file.
    """
    changed_lines = set()
    if not patch:
        return changed_lines

    # Regex to match @@ -old_start,old_count +new_start,new_count @@
    # Sometimes count is omitted if it's 1
    hunk_header_regex = re.compile(r"^@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@")

    current_new_line = 0

    lines = patch.split("\n")
    for line in lines:
        match = hunk_header_regex.match(line)
        if match:
            current_new_line = int(match.group(2))
            continue

        if line.startswith("+"):
            changed_lines.add(current_new_line)
            current_new_line += 1
        elif line.startswith(" "):
            current_new_line += 1
        elif line.startswith("-"):
            # Removed line, does not advance new_line index
            pass

    return changed_lines


def get_ai_review(filename, patch):
    """
    Sends the patch to Gemini for review.
    """
    model = genai.GenerativeModel("gemini-pro-latest")

    prompt = f"""
You are a strict Senior Go (Golang) Open Source Maintainer who follows the "Clean Code" philosophy.
Your goal is to ensure code is readable, idiomatic, and secure.

Review the following git diff for file: {filename}.

**Your Philosophy:**
We believe in **Self-Documenting Code**.
- Do NOT ask for comments to explain "what" the code is doing.
- If code is hard to understand, suggest **renaming variables** or **refactoring logic** to make it clearer, rather than suggesting adding a comment.
- Only suggest comments if there is a complex "Why" (business context) that cannot be expressed in code.

**Your Checklist:**
1. **Naming & Clarity:**
   - Flag ambiguous names (e.g., `data`, `x`, `temp`).
   - Ensure function names clearly describe their action.
   - Suggest splitting functions if they are doing too many things (Single Responsibility Principle).

2. **Go Idioms:**
   - Ensure "Exported" vs "unexported" visibility is used correctly.
   - Check that `fmt.Errorf` with `%w` is used for error wrapping.
   - Flag ignored errors (`_`) unless justified.

3. **Concurrency & Safety:**
   - Look for race conditions.
   - Ensure Mutexes are locked/unlocked correctly.

4. **Configuration:**
   - **CRITICAL:** Flag any hardcoded secrets or absolute file paths.

Output strictly valid JSON in the following format:
[
  {{
    "line": <line_number_in_new_file>,
    "message": "<your_review_comment>"
  }}
]

RULES:
- Only provide comments for lines that are ADDED or MODIFIED.
- If the code is clean and readable, return an empty list [].
- Do not include markdown formatting.

Diff:
{patch}
"""

    try:
        response = model.generate_content(prompt)
        text = response.text.strip()
        # Clean up markdown code blocks if present
        if text.startswith("```json"):
            text = text[7:]
        if text.startswith("```"):
            text = text[3:]
        if text.endswith("```"):
            text = text[:-3]

        return json.loads(text.strip())
    except Exception as e:
        print(f"Error generating/parsing review for {filename}: {e}")
        return []


def main():
    github_token = os.environ.get("GITHUB_TOKEN")
    gemini_api_key = os.environ.get("GEMINI_API_KEY")

    if not github_token or not gemini_api_key:
        print("Error: GITHUB_TOKEN and GEMINI_API_KEY are required.")
        return

    genai.configure(api_key=gemini_api_key)

    # Get context from event
    with open(os.environ["GITHUB_EVENT_PATH"], "r") as f:
        event_data = json.load(f)

    auth = Auth.Token(github_token)
    g = Github(auth=auth)

    if "pull_request" in event_data:
        pr_number = event_data["pull_request"]["number"]
        repo_name = os.environ["GITHUB_REPOSITORY"]

        repo = g.get_repo(repo_name)
        pr = repo.get_pull(pr_number)
    else:
        print("Not a pull request event. Exiting.")
        return

    print(f"Reviewing PR #{pr_number}: {pr.title}")

    commits = pr.get_commits()
    last_commit = commits[commits.totalCount - 1]

    files = pr.get_files()

    comments_to_post = []

    for file in files:
        if file.status == "removed":
            continue

        # Skip binary files or extremely large files if needed
        if not file.patch:
            print(f"Skipping {file.filename} (no patch available)")
            continue

        print(f"Analyzing {file.filename}...")

        changed_lines = parse_diff_to_changed_lines(file.patch)

        review_items = get_ai_review(file.filename, file.patch)

        for item in review_items:
            line = item.get("line")
            message = item.get("message")

            if line in changed_lines:
                print(f"  - Comment on line {line}: {message}")
                comments_to_post.append(
                    {
                        "path": file.filename,
                        "line": line,
                        "side": "RIGHT",
                        "body": message,
                    }
                )
            else:
                print(f"  - Skipped comment on line {line} (not in changed lines)")

    if comments_to_post:
        print(f"Posting {len(comments_to_post)} comments...")
        try:
            pr.create_review(
                commit=last_commit,
                body="Gemini Code Review Summary",
                event="COMMENT",
                comments=comments_to_post,
            )
            print("Review posted successfully.")
        except Exception as e:
            print(f"Failed to post review: {e}")
    else:
        print("No comments to post.")


if __name__ == "__main__":
    main()
