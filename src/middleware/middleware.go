package middleware

import (
	"judgeBackend/src/test"
	"judgeBackend/src/util/sample"
	"log"
)

type Middleware interface {
	Wrap(test *test.Test) test.Test
	String() string
}

func SelectTest(s sample.Sample) test.Test {
	var t test.Test
	switch s.Spec.Lang {
	case sample.SQLite:
		t = &test.SQLiteTest{}
	case sample.Postgres:
		t = &test.PGSQLTest{}
	default:
		log.Fatal("no default sample type")
	}
	return SelectMiddleware(t, s)
}

func SelectMiddleware(t test.Test, s sample.Sample) test.Test {
	x := t
	mwList := map[string]Middleware{"trigger": &Trigger{}}
	for k, mw := range mwList {
		_, found := s.Middleware[k]
		if found {
			x = mw.Wrap(&t)
		}
	}
	return x
}
