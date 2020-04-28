package middleware

import (
	"errors"
	"fmt"
	"judgeBackend/src/test"
	"judgeBackend/src/util"
	"judgeBackend/src/util/sample"
	"log"
)

type Trigger struct {
	wrappedTest *test.Test
	spec        map[string]string
	userInput   string
	answerInput string
}

func (t *Trigger) String() string {
	return "trigger"
}

func (t *Trigger) Run(report chan util.Report) {
	s := (*t.wrappedTest).GetSample()
	_, err := s.DB.Exec(t.userInput)
	if err != nil {
		r := util.Report{}
		r.Grade = 0
		r.Summary = fmt.Sprintf("%s is incorrect", s.Name)
		report <- r
		return
	}
	_, err = s.DB.Exec(t.answerInput)
	if err != nil {
		log.Fatal("trigger creation failed, check your yaml file")
	}
	(*t.wrappedTest).Run(report)
}

func (t *Trigger) Init(s sample.Sample, input string) error {
	spec, ok := s.Middleware["trigger"]
	if !ok {
		return errors.New("read spec for trigger middleware failed, please check your yaml file")
	}
	t.spec = spec
	t.userInput = input
	t.answerInput = s.SQL
	input = spec["userQuery"]
	s.SQL = spec["answerQuery"]
	return (*t.wrappedTest).Init(s, input)
}

func (t *Trigger) GetSample() sample.Sample {
	return (*t.wrappedTest).GetSample()
}

func (t *Trigger) Close() {
	(*t.wrappedTest).Close()
}

func (t *Trigger) Wrap(s *test.Test) test.Test {
	t.wrappedTest = s
	return t
}
