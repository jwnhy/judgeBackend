package sample

import (
	"database/sql"
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Sample struct {
	Name        string `yaml:"name"`
	Assignment  string
	Filename    string                       `yaml:"filename"`
	Description string                       `yaml:"description"`
	Content     string                       `yaml:"content"`
	Regex       string                       `yaml:"regex"`
	Rows        []interface{}                `yaml:"rows"`
	SQL         string                       `yaml:"sql"`
	Spec        Spec                         `yaml:"spec"`
	Value       float64                      `yaml:"value"`
	Middleware  map[string]map[string]string `yaml:"middleware"`
	DB          *sql.DB
	TmpFile     string
}

func LoadFromFile(filepath string) (*Sample, error) {
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
	return nil
}

func (s Sample) Tag() string {
	return s.Name + "-" + s.Assignment
}
