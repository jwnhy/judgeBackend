package main

import (
	"fmt"
	"judgeBackend/src/service"
	"log"
	_ "net/http/pprof"
	"time"
)

func main() {
	t1 := time.Now()
	err := service.Judge("Assignment3 Complex query/grades.csv", "samples/ass3")
	if err != nil {
		log.Fatal(err)
	}
	elapsed := time.Since(t1)
	fmt.Println("App elapsed: ", elapsed)
}
