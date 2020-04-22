package test

import (
	"judgeBackend/src/basestruct/report"
	pgsqlJudge "judgeBackend/src/service/pgsql-judge"
	"judgeBackend/src/service/sample"
	sqliteJudge "judgeBackend/src/service/sqlite-judge"
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
		return &sqliteJudge.SQLiteTest{}
	case sample.Postgres:
		return &pgsqlJudge.PGSQLTest{}
	default:
		log.Fatal("no default sample type")
	}
	return nil
}
