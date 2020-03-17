package sakai

import (
	"encoding/csv"
	"fmt"
	"github.com/tushar2708/altcsv"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type Student struct {
	Sid              string
	LastName         string
	FirstName        string
	Grade            float32
	SubmissionDate   string
	SubmissionStatus string
	Submissions      []*os.File
	Comment          *os.File
	Summary          []string
	FileContent      map[string]string
}

func (s Student) ToStringArray() []string {
	grade := fmt.Sprintf("%.0f", s.Grade)
	res := []string{s.Sid, s.Sid, s.LastName, s.FirstName, grade, s.SubmissionDate, s.SubmissionStatus}
	return res
}
func rowValid(row []string) bool {
	id := row[0]
	_, err := strconv.ParseInt(id, 10, 32)
	return len(row) == 7 && err == nil
}
func LoadFromCSV(gradeCSV string) ([]*Student, error) {
	gradeFile, err := os.Open(gradeCSV)
	if err != nil {
		return nil, err
	}
	baseDir := path.Dir(gradeFile.Name())
	studentReader := csv.NewReader(gradeFile)
	studentReader.LazyQuotes = true
	studentReader.FieldsPerRecord = -1
	studentRows, err := studentReader.ReadAll()
	if err != nil {
		return nil, err
	}
	res := make([]*Student, 0)
	for _, row := range studentRows {
		if !rowValid(row) {
			continue
		}
		s := &Student{row[0], row[2], row[3], 0, row[5], row[6], []*os.File{}, nil, nil, make(map[string]string)}
		commentFile := fmt.Sprintf("%s/%s, %s(%s)/comments.txt", baseDir, s.LastName, s.FirstName, s.Sid)
		s.Comment, err = os.OpenFile(commentFile, os.O_WRONLY, 777)
		if err != nil {
			return nil, err
		}
		submissionDir := fmt.Sprintf("%s/%s, %s(%s)/Submission attachment(s)", baseDir, s.LastName, s.FirstName, s.Sid)
		submissions, err := ioutil.ReadDir(submissionDir)
		if err != nil {
			return nil, err
		}
		for _, filename := range submissions {
			completeDir := fmt.Sprintf("%s/%s", submissionDir, filename.Name())
			submissionFile, err := os.Open(completeDir)
			if err != nil {
				return nil, err
			}
			s.Submissions = append(s.Submissions, submissionFile)
		}
		res = append(res, s)
	}
	return res, nil
}

func WriteToCSV(studentSlice []*Student, gradeCSV string) error {
	gradeFile, err := os.Open(gradeCSV)
	if err != nil {
		return err
	}
	studentReader := csv.NewReader(gradeFile)
	studentReader.LazyQuotes = true
	studentReader.FieldsPerRecord = -1
	firstLine, err := studentReader.Read()
	if err != nil {
		return err
	}
	secondLine, err := studentReader.Read()
	if err != nil {
		return err
	}
	thirdLine, err := studentReader.Read()
	if err != nil {
		return err
	}
	err = gradeFile.Close()
	if err != nil {
		return err
	}
	gradeFile, err = os.OpenFile(gradeCSV, os.O_WRONLY, 777)
	if err != nil {
		return err
	}
	studentWriter := altcsv.NewWriter(gradeFile)
	studentWriter.AllQuotes = true
	err = studentWriter.Write(firstLine)
	if err != nil {
		return err
	}
	err = studentWriter.Write(secondLine)
	if err != nil {
		return err
	}
	err = studentWriter.Write(thirdLine)
	if err != nil {
		return err
	}
	for _, s := range studentSlice {
		err = studentWriter.Write(s.ToStringArray())
		if err != nil {
			return err
		}
		for _, summary := range s.Summary {
			_, err = s.Comment.WriteString(summary)
			if err != nil {
				return err
			}
		}
	}
	studentWriter.Flush()
	return err
}
