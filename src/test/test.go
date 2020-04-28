package test

import (
	"judgeBackend/src/middleware"
	"judgeBackend/src/util"
	"judgeBackend/src/util/sample"
	"log"
)

type Test interface {
	Run(report chan util.Report)
	Init(s sample.Sample, input string) error
	GetSample() sample.Sample
	Close()
}

func SelectTest(s sample.Sample) Test {
	var t Test
	switch s.Spec.Lang {
	case sample.SQLite:
		t = &SQLiteTest{}
	case sample.Postgres:
		t = &PGSQLTest{}
	default:
		log.Fatal("no default sample type")
	}
	return middleware.SelectMiddleware(t, s)
}
