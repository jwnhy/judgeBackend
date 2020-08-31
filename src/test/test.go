package test

import (
	"judgeBackend/src/util"
	"judgeBackend/src/util/sample"
)

type Test interface {
	Run(report chan util.Report)
	Init(s sample.Sample, input map[string]string) error
	GetSample() sample.Sample
	Close()
}

