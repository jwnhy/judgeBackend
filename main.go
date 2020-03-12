package main

import (
	"fmt"
	sqlite_judge "judgeBackend/service/sqlite-judge"
)

func main() {
	t, err := sqlite_judge.InitTest("test", "11712009", "samples/", map[string]string{"q1":"select * from movies"})
	fmt.Println(err)
	t.RunTest()
	fmt.Print(t.Grade())
}
