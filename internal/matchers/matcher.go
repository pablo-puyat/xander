package matchers

import "xander/internal/comicvine"

type Matcher interface {
	Match(results []comicvine.Result, query string) *comicvine.Result
	Name() string
}
