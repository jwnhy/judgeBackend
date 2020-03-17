package main

import (
	"fmt"
	sqlite_judge "judgeBackend/service/sqlite-judge"
	_ "net/http/pprof"
	"time"
)

func main() {
	t1 := time.Now() // get current time
	err := sqlite_judge.Judge("Assignment2 Queries/grades.csv", "samples/Assignment1")
	fmt.Println(err)
	elapsed := time.Since(t1)
	fmt.Println("App elapsed: ", elapsed)
}
