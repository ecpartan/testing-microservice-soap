package main

import (
	"fmt"
	"log"
	"os"

	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/testing-microservice-soap/config"
	"github.com/ecpartan/testing-microservice-soap/report"
	"github.com/ecpartan/testing-microservice-soap/test"
	"github.com/ecpartan/testing-microservice-soap/test/updatefw"
	"github.com/robfig/cron/v3"
)

func main() {
	c := cron.New()
	config.LoadConfig()
	logger.InitLogger(os.Stdout)

	updatefw.GetUpdatefw()
	cfg := config.GetServerConfig()
	cfgfw := config.GetFirmwareConfig()

	var url string
	if url = os.Getenv("SERVER_URL"); url == "" {
		url = fmt.Sprintf("http://%s:%d/integral", cfg.Host, cfg.Port)
	}

	if len(os.Args) > 1 {
		logger.LogDebug("Starting server...", len(os.Args))

		switch os.Args[1] {
		case "test":
			logger.LogDebug("Starting tests...", os.Args[1])

			rep := test.Execute_test(cfg, cfgfw, url)

			fmt.Println(rep)
			report.SendReportMail(rep)
		case "onetest":
			logger.LogDebug("Starting onetest...%s", os.Args[1])
			if err := test.Execute_one_script(cfgfw, os.Args[2], url); err != nil {
				logger.LogDebug("Test failed", os.Args[2], err)
			} else {
				logger.LogDebug("Test success", os.Args[2])
			}
		}
		os.Exit(1)
	}

	info := config.GetServerConfig()
	cronstr := fmt.Sprintf("%d %d * * *", info.CronMinute, info.CronHour)
	logger.LogDebug("Starting server...", cronstr)

	_, err := c.AddFunc(cronstr, func() {
		rep := test.Execute_test(cfg, cfgfw, url)

		fmt.Println(rep)
		report.SendReportMail(rep)
	})

	if err != nil {
		log.Fatal(err)
	}

	c.Start()

	select {}
}
