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
	err := service.Judge("Assignment4 Trigger Final/grades.csv", "samples/ass4")
	if err != nil {
		log.Fatal(err)
	}
	elapsed := time.Since(t1)
	fmt.Println("App elapsed: ", elapsed)
}
