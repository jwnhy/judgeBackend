package test

import (
	"judgeBackend/basestruct/report"
	"judgeBackend/service/sample"
	sqlite_judge "judgeBackend/service/sqlite-judge"
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
	default:
		log.Fatal("no default sample type")
	}
	return nil
}
