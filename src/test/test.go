package test

import (
	"judgeBackend/src/util"
	"judgeBackend/src/util/sample"
	"log"
)

type Test interface {
	Run(report chan util.Report)
	Init(s sample.Sample, input string) error
	Close()
}

func SelectTest(s sample.Sample) Test {
	switch s.Spec.Lang {
	case sample.SQLite:
		return &SQLiteTest{}
	case sample.Postgres:
		return &PGSQLTest{}
	default:
		log.Fatal("no default sample type")
	}
	return nil
}
