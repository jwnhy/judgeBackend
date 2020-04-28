package test

import (
	"database/sql"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"io/ioutil"
	"judgeBackend/src/util"
	"judgeBackend/src/util/sample"
	"log"
	"os"
)

var sqliteCache = util.NewCache()

type SQLiteTest struct {
	input  string
	sample sample.Sample
}

func (t *SQLiteTest) String() string {
	return "SQLiteTest"
}

func (t *SQLiteTest) GetSample() sample.Sample {
	return t.sample
}

func (t *SQLiteTest) Init(s sample.Sample, input string) error {
	if s.Spec.Lang == sample.SQLite {
		sourceDB, err := os.Open(s.Spec.Database)
		if err != nil {
			return err
		}
		tmpDB, err := ioutil.TempFile("/tmp", "judge")
		if err != nil {
			return err
		}
		_, err = io.Copy(tmpDB, sourceDB)
		if err != nil {
			return err
		}
		s.DB, err = sql.Open("sqlite3", tmpDB.Name())
		s.TmpFile = tmpDB.Name()
		if err != nil {
			log.Fatal(err)
		}
		*t = SQLiteTest{input, s}
	}
	return nil
}

func (t *SQLiteTest) Run(reportChan chan util.Report) {
	s := t.sample
	var standardSlice []interface{}
	r := &util.Report{}
	res, ok := sqliteCache.Get(s.Name, s.SQL)
	if !ok {
		standardRows, err := s.DB.Query(s.SQL)
		if err != nil {
			log.Println(s.SQL)
			log.Fatal(err)
		}
		standardSlice, _, err = util.ScanInterface(standardRows)
		if err != nil {
			log.Fatal(err)
		}
		sqliteCache.Set(s.Name, s.SQL, standardSlice)
	} else {
		standardSlice = res
	}
	userRows, err := s.DB.Query(t.input)
	if err != nil {
		r.Grade = 0
		r.Summary = err.Error() + "\n"
		goto SEND
	} else {
		userSlice, _, err := util.ScanInterface(userRows)
		if err != nil {
			r.Grade = 0
			r.Summary = err.Error() + "\n"
			goto SEND

		}
		if len(userSlice) != len(standardSlice) {
			r.Grade = 0
			r.Summary = fmt.Sprintf("%s is wrong, row number is wrong\n", s.Name)
			goto SEND
		}
		if s.Spec.IsSet {
			s1 := mapset.NewSetFromSlice(standardSlice)
			s2 := mapset.NewSetFromSlice(userSlice)
			if !s1.Equal(s2) {
				r.Grade = 0
				r.Summary = fmt.Sprintf("%s is wrong\n", s.Name)
				goto SEND
			}
		} else {
			for i, s1 := range standardSlice {
				s2 := userSlice[i]
				if s1 != s2 {
					r.Grade = 0
					r.Summary = fmt.Sprintf("%s is wrong\n", s.Name)
					goto SEND
				}
			}
		}
	}
	r.Grade = s.Value
	r.Summary = fmt.Sprintf("%s is correct\n", s.Name)
SEND:
	reportChan <- *r
}

func (t SQLiteTest) Close() {
	err := os.Remove(t.sample.TmpFile)
	if err != nil {
		fmt.Println(err)
	}
}
