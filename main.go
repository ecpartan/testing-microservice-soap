package main

import (
	"fmt"
	"log"
	"os"

	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/testing-microservice-soap/config"
	"github.com/ecpartan/testing-microservice-soap/report"
	"github.com/ecpartan/testing-microservice-soap/test"
	"github.com/robfig/cron/v3"
)

func main() {

	c := cron.New()
	config.LoadConfig()
	logger.InitLogger(os.Stdout)

	if len(os.Args) > 1 {
		logger.LogDebug("Starting server...", len(os.Args))

		if os.Args[1] == "test" {
			logger.LogDebug("Starting server...", os.Args[1])

			rep := test.Execute_test()

			fmt.Println(rep)
			report.SendReportMail(rep)
		}
		os.Exit(1)
	}

	info := config.GetServerConfig()
	cronstr := fmt.Sprintf("%d %d * * *", info.CronMinute, info.CronHour)
	logger.LogDebug("Starting server...", cronstr)

	_, err := c.AddFunc(cronstr, func() {
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
