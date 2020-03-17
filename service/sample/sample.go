package sample

import (
	"database/sql"
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Sample struct {
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	Content     string        `yaml:"content"`
	Regex       string        `yaml:"regex"`
	Rows        []interface{} `yaml:"rows"`
	SQL         string        `yaml:"sql"`
	Spec        Spec          `yaml:"spec"`
	Value       float32       `yaml:"value"`
	DB          *sql.DB
	TmpFile     string
}

func ReadFromFile(filepath string) (*Sample, error) {
	f, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	sample := Sample{}
	err = yaml.Unmarshal(f, &sample)
	if err != nil {
		return nil, err
	}
	err = sample.CheckConsistency()
	return &sample, err
}

func (s Sample) CheckConsistency() error {
	spec := s.Spec
	if spec.Type == Regex && s.Regex == "" {
		return errors.New("regular expression cannot be empty")
	}
	if spec.Type == Row && len(s.Rows) == 0 {
		return errors.New("row array cannot be empty")
	}
	return nil
}
