package sqlite_judge

import (
	"context"
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
	"path"
)

var tmpDir string

type SQLiteTest struct {
	input   map[string]string
	samples []sample.Sample
	summary []string
	grade   float32
}

func (t *SQLiteTest) InitTest(dirPath string, input map[string]string) error {
	*t = SQLiteTest{input, []sample.Sample{}, []string{}, 0}
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}
	if tmpDir == "" {
		tmpDir, err = ioutil.TempDir("/tmp", "")
		if err != nil {
			return err
		}
	}
	for _, testcase := range files {
		testcasePath := path.Join(dirPath, testcase.Name())
		s, err := sample.ReadFromFile(testcasePath)
		if err != nil {
			return err
		}
		if s.Spec.Lang == sample.SQLite {
			sourceDB, err := os.Open(s.Spec.Database)
			if err != nil {
				return err
			}

			tmpDB, err := ioutil.TempFile(tmpDir, "judge")
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
			t.samples = append(t.samples, *s)
		}
	}
	return nil
}

func (t *SQLiteTest) RunTest(grade chan float32, summary chan []string) {
	for _, s := range t.samples {
		if t.input[s.Name] == "" {
			continue
		}
		standardRows, err := s.DB.QueryContext(context.Background(), s.SQL)
		if err != nil {
			log.Println(s.SQL)
			log.Fatal(err)
		}
		standardSlice, err := util.ScanInterface(standardRows)
		if err != nil {
			log.Fatal(err)
		}
		userRows, err := s.DB.QueryContext(context.Background(), t.input[s.Name])
		if err != nil {
			continue
		}
		userSlice, err := util.ScanInterface(userRows)
		if err != nil {
			log.Fatal(err)
		}
		if s.Spec.IsSet {
			s1 := mapset.NewSetFromSlice(standardSlice)
			s2 := mapset.NewSetFromSlice(userSlice)
			if !s1.Equal(s2) {
				goto WRONG
			}
		} else {
			if len(standardSlice) == len(userSlice) {
				for i, s1 := range standardSlice {
					s2 := userSlice[i]
					if s1 != s2 {
						goto WRONG
					}
				}
			}
		}
		t.summary = append(t.summary, fmt.Sprintf("%s is Correct\n", s.Name))
		t.grade += s.Value
		continue
	WRONG:
		t.summary = append(t.summary, fmt.Sprintf("%s is Wrong\n", s.Name))
		continue
	}
	grade <- t.grade
	summary <- t.summary
}
func (t SQLiteTest) Grade() float32 {
	return t.grade
}

func (t SQLiteTest) Close() {
	_ = os.RemoveAll(tmpDir)
	_ = os.Remove(tmpDir)
}
