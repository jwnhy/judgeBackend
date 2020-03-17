package base

type Test interface {
	InitTest(dirPath string, input map[string]string) error
	RunTest(grade chan float32, summary chan []string)
}
