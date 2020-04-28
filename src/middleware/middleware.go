package middleware

import (
	"judgeBackend/src/test"
	"judgeBackend/src/util/sample"
)

type Middleware interface {
	Wrap(test *test.Test) *test.Test
}

func SelectMiddleware(s sample.Sample) []Middleware {
	return nil
}
