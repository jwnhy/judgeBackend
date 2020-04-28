package test

import (
	"database/sql"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"github.com/docker/go-connections/nat"
	_ "github.com/lib/pq"
	"judgeBackend/src/util"
	"judgeBackend/src/util/sample"
	"log"
	"time"
)

var pgCache = util.NewCache()

const (
	WaitDuration = 3
)

type PGSQLTest struct {
	input    string
	sample   sample.Sample
	dockerId string
}

func (t PGSQLTest) String() string {
	return "PGSQLTest"
}
func (t *PGSQLTest) GetSample() sample.Sample {
	return t.sample
}
func (t *PGSQLTest) Init(s sample.Sample, input string) error {
	if s.Spec.Lang == sample.Postgres {
		built, building, err := util.ImageExist(s)
		if err != nil {
			return err
		}
		if !building && !built {
			err := util.Build(s)
			if err != nil {
				return err
			}
		}
		for !built {
			built, _, err = util.ImageExist(s)
			if err != nil {
				return err
			}
			time.Sleep(WaitDuration * time.Second)
		}
		id, err := util.StartContainer(s, []nat.Port{"5432"})
		if err != nil {
			return err
		}
		ip, err := util.GetIPAddress(id)
		connStr := fmt.Sprintf("postgres://judge:judge@%s/judge?sslmode=disable", ip)
		s.DB, err = sql.Open("postgres", connStr)
		if err = s.DB.Ping(); err != nil {
			for s.DB.Ping() != nil {
				time.Sleep(WaitDuration * time.Second)
			}
		}
		*t = PGSQLTest{input, s, id}
	}
	return nil
}

func (t *PGSQLTest) Run(reportChan chan util.Report) {
	s := t.sample
	var standardSlice []interface{}
	r := util.Report{}
	res, ok := pgCache.Get(s.Name, s.SQL)
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
		pgCache.Set(s.Name, s.SQL, standardSlice)
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
	reportChan <- r
}

func (t *PGSQLTest) Close() {
	err := util.RemoveContainer(t.dockerId)
	if err != nil {
		fmt.Println(err)
	}
}
