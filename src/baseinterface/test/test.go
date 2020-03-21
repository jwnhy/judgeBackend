package test

import (
	"judgeBackend/src/basestruct/report"
	pgsql_judge "judgeBackend/src/service/pgsql-judge"
	"judgeBackend/src/service/sample"
	sqlite_judge "judgeBackend/src/service/sqlite-judge"
	"log"
)

type Test interface {
	Run(report chan report.Report)
	Init(s sample.Sample, input string) error
	Close()
}

func SelectTest(s sample.Sample) Test {
	switch s.Spec.Lang {
	case sample.SQLite:
		return &sqlite_judge.SQLiteTest{}
	case sample.Postgres:
		return &pgsql_judge.PGSQLTest{}
	default:
		log.Fatal("no default sample type")
	}
	return nil
}
