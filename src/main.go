package main

import (
	"fmt"
	"judgeBackend/src/service/sakai"
	_ "net/http/pprof"
	"time"
)

func main() {
	t1 := time.Now() // get current time
	err := sakai.Judge("Assignment3 Complex query/grades.csv", "samples/Assignment3")
	fmt.Println(err)
	elapsed := time.Since(t1)
	fmt.Println("App elapsed: ", elapsed)
}
