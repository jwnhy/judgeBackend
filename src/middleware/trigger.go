package middleware

import (
	"judgeBackend/src/test"
	"judgeBackend/src/util"
	"judgeBackend/src/util/sample"
)

type Trigger struct {
	wrappedTest *test.Test
}

func (t *Trigger) Run(report chan util.Report) {
	t.wrappedTest.Run(report)
}

func (t *Trigger) Init(s sample.Sample, input string) error {
	return t.wrappedTest.Init(s, input)
}
func (t *Trigger) Close() {

}

func (t *Trigger) Wrap(test *test.Test) *test.Test {
	return nil
}
