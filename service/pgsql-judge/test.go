package pgsql_judge

import (
	"database/sql"
	"io/ioutil"
	"judgeBackend/base"
	"judgeBackend/service/sample"
	"path"
)

type PGSQLTest struct {
	input   map[string]string
	samples []sample.Sample
	summary []string
	grade   float32
	db      *sql.DB
}

func (t *PGSQLTest) New() base.Test {
	return &PGSQLTest{}
}

func (t *PGSQLTest) Grade() float32 {
	return t.grade
}

func (t *PGSQLTest) Init(dirPath string, input string) error {
	*t = PGSQLTest{input, []sample.Sample{}, []string{}, 0, nil}
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}
	for _, testcase := range files {
		testcasePath := path.Join(dirPath, testcase.Name())
		s, err := sample.LoadFromFile(testcasePath)
		if err != nil {
			return err
		}
		t.samples = append(t.samples, *s)
	}
	return nil
}

func (t *PGSQLTest) Run(grade chan float32, summary chan []string) {
	panic("implement me")
}

func (t *PGSQLTest) Close() {
	panic("implement me")
}
