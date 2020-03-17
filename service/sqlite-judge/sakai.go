package sqlite_judge

import (
	"judgeBackend/service/sakai"
	"log"
	"path"
	"strings"
)

func StudentToInput(studentSlice []*sakai.Student) error {
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

func InitAndRun(t Test, sampleDir string, input map[string]string, gradeChan chan float32, summaryChan chan []string) {
	err := t.InitTest(sampleDir, input)
	if err != nil {
		log.Fatal(err)
	}
	t.RunTest(gradeChan, summaryChan)
	defer t.Close()
}

func Judge(gradeCSV, sampleDir string) error {
	studentSlice, err := sakai.LoadFromCSV(gradeCSV)
	if err != nil {
		return err
	}
	err = StudentToInput(studentSlice)
	if err != nil {
		return err
	}
	gradeChan := make([]chan float32, len(studentSlice))
	summaryChan := make([]chan []string, len(studentSlice))
	for i, student := range studentSlice {
		gradeChan[i] = make(chan float32)
		summaryChan[i] = make(chan []string)
		go InitAndRun(sampleDir, student.FileContent, gradeChan[i], summaryChan[i])
	}
	for i, student := range studentSlice {
		student.Grade = <-gradeChan[i]
		student.Summary = <-summaryChan[i]
	}
	err = sakai.WriteToCSV(studentSlice, gradeCSV)
	if err != nil {
		return err
	}
	return nil
}
