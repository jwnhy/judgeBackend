package service

import (
	"encoding/csv"
	"fmt"
	"github.com/dimchansky/utfbom"
	"github.com/tushar2708/altcsv"
	"io/ioutil"
	"judgeBackend/src/test"
	"judgeBackend/src/util"
	"judgeBackend/src/util/sample"
	"log"
	"os"
	"path"
	"strconv"
	"time"
)

type Student struct {
	SID              string
	LastName         string
	FirstName        string
	Grade            float64
	SubmissionDate   string
	SubmissionStatus string
	Submissions      []*os.File
	Comment          *os.File
	Summary          []string
	FileContent      map[string]string
}

func (s Student) ToStringArray() []string {
	grade := fmt.Sprintf("%.0f", s.Grade)
	res := []string{s.SID, s.SID, s.LastName, s.FirstName, grade, s.SubmissionDate, s.SubmissionStatus}
	return res
}
func rowValid(row []string) bool {
	id := row[0]
	_, err := strconv.ParseInt(id, 10, 32)
	return len(row) == 7 && err == nil
}
func LoadFromCSV(gradeCSV string) (map[string]*Student, error) {
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
	res := make(map[string]*Student)
	for _, row := range studentRows {
		if !rowValid(row) {
			continue
		}
		s := &Student{row[0], row[2], row[3], 0, row[5], row[6], []*os.File{}, nil, nil, make(map[string]string)}
		commentFile := fmt.Sprintf("%s/%s, %s(%s)/comments.txt", baseDir, s.LastName, s.FirstName, s.SID)
		s.Comment, err = os.OpenFile(commentFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return nil, err
		}
		submissionDir := fmt.Sprintf("%s/%s, %s(%s)/Submission attachment(s)", baseDir, s.LastName, s.FirstName, s.SID)
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
		res[s.SID] = s
	}
	return res, nil
}

func WriteToCSV(studentSlice map[string]*Student, gradeCSV string) error {
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
	gradeFile, err = os.OpenFile(gradeCSV, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
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

func StudentToInput(studentSlice map[string]*Student) error {
	for _, student := range studentSlice {
		res := make(map[string]string, 0)
		for _, f := range student.Submissions {
			filename := path.Base(f.Name())
			f, err := os.Open(f.Name())
			if err != nil {
				return err
			}
			byteContent, err := ioutil.ReadAll(utfbom.SkipOnly(f))
			if err != nil {
				return err
			}
			res[filename] = string(byteContent)
		}
		student.FileContent = res
	}
	return nil
}

func InitAndRun(sampleDir string, student *Student, reportChan chan util.Report) {
	files, err := ioutil.ReadDir(sampleDir)
	if err != nil {
		log.Fatal(err)
	}
	input := student.FileContent
	tmpReportChan := make(chan util.Report)
	for _, f := range files {
		s, err := sample.LoadFromFile(path.Join(sampleDir, f.Name()))
		if err != nil {
			log.Fatal(err)
		}
		t := test.SelectTest(*s)

		if input[s.Filename] == "" {
			reportChan <- util.Report{SID: student.SID, Grade: 0, Summary: s.Filename + " not found.\n", End: false}
			continue
		}

		err = t.Init(*s, input[s.Filename])
		if err != nil {
			log.Fatal(err)
		}
		go t.Run(tmpReportChan)
		r := <-tmpReportChan
		r.SID = student.SID
		r.End = false
		reportChan <- r
		t.Close()
	}
	r := util.Report{SID: student.SID, Summary: time.Now().String(), End: true}
	reportChan <- r
}

func Judge(gradeCSV, sampleDir string) error {
	studentMap, err := LoadFromCSV(gradeCSV)
	tmpMap := make(map[string]*Student)
	for k, _ := range studentMap {
		tmpMap[k] = nil
	}
	if err != nil {
		return err
	}

	err = StudentToInput(studentMap)
	if err != nil {
		return err
	}
	reportChanQueue := make(chan util.Report, len(studentMap))
	for _, student := range studentMap {
		go InitAndRun(sampleDir, student, reportChanQueue)
	}
	for len(tmpMap) > 0 {
		r := <-reportChanQueue
		if r.End {
			delete(tmpMap, r.SID)
			studentMap[r.SID].Summary = append(studentMap[r.SID].Summary, r.Summary)
			fmt.Printf("Finished %s with grade %.2f\n", r.SID, studentMap[r.SID].Grade)
		} else {
			studentMap[r.SID].Grade += r.Grade
			studentMap[r.SID].Summary = append(studentMap[r.SID].Summary, r.Summary)
		}
	}
	err = WriteToCSV(studentMap, gradeCSV)
	if err != nil {
		return err
	}
	return nil
}
