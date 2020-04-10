package test

import (
	"fmt"
	"judgeBackend/src/service/sakai"
	"testing"
	"time"
)

func TestJudge(t *testing.T) {
	t1 := time.Now() // get current time
	err := sakai.Judge("Assignment2 Queries/grades.csv", "samples/Assignment2")
	fmt.Println(err)
	elapsed := time.Since(t1)
	fmt.Println("App elapsed: ", elapsed)
}
