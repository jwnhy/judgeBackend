package pgsql_judge

import (
	"database/sql"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"github.com/docker/go-connections/nat"
	_ "github.com/lib/pq"
	"judgeBackend/src/basestruct/report"
	"judgeBackend/src/basestruct/sqlcache"
	"judgeBackend/src/service/docker"
	"judgeBackend/src/service/sample"
	"judgeBackend/src/util"
	"log"
	"time"
)

var sqlCache = sqlcache.New()

const (
	WaitDuration = 3
)

type PGSQLTest struct {
	input    string
	sample   sample.Sample
	dockerId string
}

func (t *PGSQLTest) Init(s sample.Sample, input string) error {
	if s.Spec.Lang == sample.Postgres {
		built, building, err := docker.ImageExist(s)
		if err != nil {
			return err
		}
		if !building && !built {
			err := docker.Build(s)
			if err != nil {
				return err
			}
		}
		for !built {
			built, _, err = docker.ImageExist(s)
			if err != nil {
				return err
			}
			time.Sleep(WaitDuration * time.Second)
		}
		id, err := docker.StartContainer(s, []nat.Port{"5432"})
		if err != nil {
			return err
		}
		ip, err := docker.GetIPAddress(id)
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

func (t *PGSQLTest) Run(reportChan chan report.Report) {
	s := t.sample
	var standardSlice []interface{}
	r := &report.Report{}
	res, ok := sqlCache.Get(s.Name, s.SQL)
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
		sqlCache.Set(s.Name, s.SQL, standardSlice)
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

func (t *PGSQLTest) Close() {
	err := docker.RemoveContainer(t.dockerId)
	if err != nil {
		fmt.Println(err)
	}
}
