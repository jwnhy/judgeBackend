package sqlite_judge

import (
	"database/sql"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"io/ioutil"
	"judgeBackend/service/sample"
	"judgeBackend/util"
	"log"
	"os"
)

type SQLiteTest struct {
	input  string
	sample sample.Sample
}

func (t *SQLiteTest) Init(dir string, s sample.Sample, input string) error {
	if s.Spec.Lang == sample.SQLite {
		sourceDB, err := os.Open(s.Spec.Database)
		if err != nil {
			return err
		}
		tmpDB, err := ioutil.TempFile(dir, "judge")
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

func (t *SQLiteTest) Run(grade chan float64, summary chan string) {
	s := t.sample
	standardRows, err := s.DB.Query(s.SQL)
	if err != nil {
		log.Println(s.SQL)
		log.Fatal(err)
	}
	standardSlice, err := util.ScanInterface(standardRows)
	if err != nil {
		log.Fatal(err)
	}
	userRows, err := s.DB.Query(t.input)
	if err != nil {
		grade <- 0
		summary <- err.Error() + "\n"
		return
	} else {
		userSlice, err := util.ScanInterface(userRows)
		if err != nil {
			grade <- 0
			summary <- err.Error() + "\n"
		}
		if s.Spec.IsSet {
			s1 := mapset.NewSetFromSlice(standardSlice)
			s2 := mapset.NewSetFromSlice(userSlice)
			if !s1.Equal(s2) {
				grade <- 0
				summary <- fmt.Sprintf("%s is wrong\n", s.Name)
				return
			}
		} else {
			if len(standardSlice) == len(userSlice) {
				for i, s1 := range standardSlice {
					s2 := userSlice[i]
					if s1 != s2 {
						grade <- 0
						summary <- fmt.Sprintf("%s is wrong\n", s.Name)
						return
					}
				}
			}
		}
	}
	grade <- s.Value
	summary <- fmt.Sprintf("%s is correct\n", s.Name)
	return
}

func (t SQLiteTest) Close() {
	err := os.Remove(t.sample.TmpFile)
	if err != nil {
		fmt.Println(err)
	}
}
