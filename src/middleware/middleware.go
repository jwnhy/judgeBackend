package middleware

import (
	"judgeBackend/src/test"
	"judgeBackend/src/util/sample"
)

type Middleware interface {
	Wrap(test *test.Test) test.Test
	String() string
}

func SelectMiddleware(t test.Test, s sample.Sample) test.Test {
	mwList := map[string]Middleware{"trigger": &Trigger{}}
	for k, mw := range mwList {
		_, found := s.Middleware[k]
		if found {
			t = mw.Wrap(&t)
		}
	}
	return t
}
