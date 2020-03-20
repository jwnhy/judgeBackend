package main

import (
	"fmt"
	"judgeBackend/service/sakai"
	_ "net/http/pprof"
	"time"
)

func main() {
	t1 := time.Now() // get current time
	err := sakai.Judge("Assignment2 Queries/grades.csv", "samples/Assignment1")
	fmt.Println(err)
	elapsed := time.Since(t1)
	fmt.Println("App elapsed: ", elapsed)
}
