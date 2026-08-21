package main

import (
	"fmt"
	"log"
	"os"

	"github.com/robfig/cron/v3"

	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/testing-microservice-soap/config"
	"github.com/ecpartan/testing-microservice-soap/report"
	"github.com/ecpartan/testing-microservice-soap/test"
)

func main() {

	c := cron.New()
	config.LoadConfig()
	logger.InitLogger(os.Stdout)

	_, err := c.AddFunc("52 11 * * *", func() {
		rep := test.Execute_test()

		fmt.Println(rep)
		report.SendReportMail(rep)
	})

	if err != nil {
		log.Fatal(err)
	}

	c.Start()

	select {}
}
