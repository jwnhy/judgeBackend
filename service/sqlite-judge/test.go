package sqlite_judge

import (
	"context"
	"database/sql"
	mapset "github.com/deckarep/golang-set"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"io/ioutil"
	"judgeBackend/service/sample"
	"judgeBackend/util"
	"log"
	"os"
)

type Test struct {
	username string
	sid      string
	input    map[string]string
	samples  []sample.Sample
	grade    float32
}

func InitTest(username, sid, dirPath string, input map[string]string) (*Test, error) {
	t := Test{username, sid, input, []sample.Sample{}, 0}
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	for _, testcase := range files {
		s, err := sample.ReadFromFile(dirPath + testcase.Name())
		if err != nil {
			return nil, err
		}
		if s.Spec.Lang == sample.SQLite {
			sourceDB, err := os.Open(s.Spec.Database)
			if err != nil {
				return nil, err
			}
			tmpDB, err := ioutil.TempFile("/tmp", "")
			if err != nil {
				return nil, err
			}
			_, err = io.Copy(tmpDB, sourceDB)
			if err != nil {
				return nil, err
			}
			s.DB, err = sql.Open("sqlite3", tmpDB.Name())
			if err != nil {
				log.Fatal(err)
			}
			t.samples = append(t.samples, *s)
		}
	}
	return &t, nil
}

func (t *Test) RunTest() {
	for _, s := range t.samples {
		standardRows, err := s.DB.QueryContext(context.Background(), s.SQL)
		if err != nil {
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
			if s1.Equal(s2) {
				t.grade += s.Value
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
			t.grade += s.Value
		WRONG:
			continue
		}
	}
}
func (t Test) Grade()float32 {
	return t.grade
}
