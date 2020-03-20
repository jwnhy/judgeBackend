package sakai

import (
	"encoding/csv"
	"fmt"
	"github.com/tushar2708/altcsv"
	"io/ioutil"
	"judgeBackend/base"
	"judgeBackend/service/sample"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

type Student struct {
	Sid              string
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

func StudentToInput(studentSlice []*Student) error {
	for _, student := range studentSlice {
		res := make(map[string]string, 0)
		for _, f := range student.Submissions {
			filename := strings.TrimSuffix(path.Base(f.Name()), path.Ext(f.Name()))
			byteContent := make([]byte, 1024*1024)
			_, err := f.Read(byteContent)
			if err != nil {
				return err
			}
			res[filename] = string(byteContent)
		}
		student.FileContent = res
	}
	return nil
}

func InitAndRun(sampleDir string, input map[string]string, gradeChan chan float64, summaryChan *[]chan string) {
	files, err := ioutil.ReadDir(sampleDir)
	if err != nil {
		log.Fatal(err)
	}
	*summaryChan = make([]chan string, len(files))
	gradeChanList := make([]chan float64, len(files))
	grade := float64(0)
	tmpDir, err := ioutil.TempDir("/tmp", "")
	if err != nil {
		log.Fatal(err)
	}
	for i, f := range files {
		gradeChanList[i] = make(chan float64)
		(*summaryChan)[i] = make(chan string)
		s, err := sample.LoadFromFile(path.Join(sampleDir, f.Name()))
		if err != nil {
			log.Fatal(err)
		}
		t := base.SelectTest(*s)
		if input[s.Name] == "" {
			continue
		}

		err = t.Init(tmpDir, *s, input[s.Name])
		if err != nil {
			log.Fatal(err)
		}
		go t.Run(gradeChanList[i], (*summaryChan)[i])
	}
	for _, c := range gradeChanList {
		grade += <-c
	}
	err = os.RemoveAll(tmpDir)
	if err != nil {
		fmt.Println(err)
	}
	gradeChan <- grade
}

func Judge(gradeCSV, sampleDir string) error {
	studentSlice, err := LoadFromCSV(gradeCSV)
	if err != nil {
		return err
	}
	err = StudentToInput(studentSlice)
	if err != nil {
		return err
	}
	gradeChan := make([]chan float64, len(studentSlice))
	summaryChan := make([][]chan string, len(studentSlice))
	for i, student := range studentSlice {
		gradeChan[i] = make(chan float64)
		go InitAndRun(sampleDir, student.FileContent, gradeChan[i], &summaryChan[i])
	}
	for i, student := range studentSlice {
		student.Grade = <-gradeChan[i]
		for _, summary := range summaryChan[i] {
			student.Summary = append(student.Summary, <-summary)
		}
	}
	err = WriteToCSV(studentSlice, gradeCSV)
	if err != nil {
		return err
	}
	return nil
}
