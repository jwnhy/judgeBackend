package main

import (
	"fmt"
	"github.com/dimchansky/utfbom"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"judgeBackend/src/baseinterface/test"
	"judgeBackend/src/basestruct/report"
	"judgeBackend/src/service/sample"
	"log"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "InfinityJudge"
	app.Usage = "Input file, output grade."
	app.Version = "0.0.1cc"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "sample, s",
			Value: "sample.yaml",
			Usage: "read sample from `FILE`",
		}, cli.StringFlag{
			Name:  "query, q",
			Value: "query.sql",
			Usage: "read user query from `FILE`",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "judge",
			Aliases: []string{"j"},
			Usage:   "judge based on sample and input, result will be send to stdout",
			Action: func(c *cli.Context) error {
				s, err := sample.LoadFromFile(c.String("sample"))
				if err != nil {
					log.Fatal(err)
				}
				t := test.SelectTest(*s)
				f, err := os.Open(c.String("query"))
				if err != nil {
					log.Fatal(err)
				}
				byteContent, err := ioutil.ReadAll(utfbom.SkipOnly(f))
				err = t.Init(*s, string(byteContent))
				if err != nil {
					log.Fatal(err)
				}
				reportChan := make(chan report.Report)
				go t.Run(reportChan)
				r := <-reportChan
				fmt.Println(r.SID)
				fmt.Println(r.Summary)
				fmt.Println(r.Grade)
				return nil
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
