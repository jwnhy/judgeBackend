package middleware

import (
	"database/sql"
	"errors"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"judgeBackend/src/test"
	"judgeBackend/src/util"
	"judgeBackend/src/util/sample"
	"log"
	"math"
)

type Trigger struct {
	wrappedTest *test.Test
	spec        map[string][]string
	input       map[string]string
}

func (t *Trigger) String() string {
	return "trigger"
}

func (t *Trigger) Run(report chan util.Report) {
	s := (*t.wrappedTest).GetSample()
	r := util.Report{}
	db := s.DB
	for _, stmt := range t.spec["answer"] {
		_, err := db.Exec(stmt)
		if err != nil {
			log.Fatal("failed with answer, please check your yaml file")
		}
	}
	answerRes, answerQueryRes := execHelper(t.spec["test"], t.spec["query"][0], s.DB)
	for _, cleanStmt := range t.spec["clean"] {
		_, err := db.Exec(cleanStmt)
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, file := range t.spec["user"] {
		stmt, ok := t.input[file]
		if !ok {
			r.Grade = 0
			r.Summary = fmt.Sprintf("file %s not found", file)
			report <- r
			return
		}
		_, err := db.Exec(stmt)
		if err != nil {
			r.Grade = 0
			r.Summary = fmt.Sprintf("statement failed with %s", err.Error())
			report <- r
			return
		}
	}
	userRes, userQueryRes := execHelper(t.spec["test"], t.spec["query"][0], s.DB)
	var TP, FN, FP, TN float64
	for i, ans := range answerRes {
		user := userRes[i]
		if ans {
			if user {
				TP += 1
			} else {
				FN += 1
			}
		} else {
			if user {
				FP += 1
			} else {
				TN += 1
			}
		}
	}
	r.Grade = math.Max(((TP + TN) -  FP - 32.0/18.0 * FN)/(TP + TN + FP + FN) * 100, 0)
	factor := 1.0
	if int(TP + TN) == int(TP+TN+FP+FN) {
		if !mapset.NewSetFromSlice(userQueryRes).Equal(mapset.NewSetFromSlice(answerQueryRes)) {
			factor = 0.8
		}
	}
	r.Grade = r.Grade * factor
	r.Summary = fmt.Sprintf("TP: %.2f\nFP: %.2f\nTN: %.2f\nFN: %.2f\n", TP, FP, TN, FN)
	r.Summary += fmt.Sprintf("Your score will calculated based on these attributes\n")
	r.Summary += fmt.Sprintf("Current formula is ((TP + TN) - 16/9 * FP - FN)/(TP + TN + FP + FN) * factor * 100\n")
	r.Summary += fmt.Sprintf("i.e we give more punishment to these trigger which view invalid id as valid\n")
	r.Summary += fmt.Sprintf("If your trigger correctly inserted all rows, but the result is different than answer(birthday/address), your score will be multiply by a factor(0.8)\n")
	report <- r
}

func (t *Trigger) Init(s sample.Sample, input map[string]string) error {
	spec, ok := s.Middleware["trigger"]
	if !ok {
		return errors.New("read spec for trigger middleware failed, please check your yaml file")
	}
	t.spec = spec
	t.input = input
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

func execHelper(inserts []string, query string, db *sql.DB) (res []bool, row []interface{}) {
	for _, insert := range inserts {
		tx, err := db.Begin()
		if err != nil {
			log.Fatal(err)
		}
		_, err = tx.Exec(insert)
		res = append(res, err == nil)
		if err != nil {
			err = tx.Rollback()
		} else {
			err = tx.Commit()
		}

		if err != nil {
			log.Fatal(err)
		}

	}
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("result fetch failed in trigger")
	}
	rs, _, err := util.ScanInterface(rows)
	return res, rs
}
