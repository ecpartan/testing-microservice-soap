package main

import (
	"fmt"
	"os"

	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/testing-microservice-soap/config"
	"github.com/ecpartan/testing-microservice-soap/report"
	"github.com/ecpartan/testing-microservice-soap/test"
)

func main() {

	//c := cron.New()
	config.LoadConfig()
	logger.InitLogger(os.Stdout)

	rep := test.Execute_test()

	fmt.Println(rep)
	report.SendReportMail(rep)
	/*
		_, err := c.AddFunc("28 18 * * *", func() {
			rep := test.Execute_test()

			fmt.Println(rep)
			report.SendReportMail(rep)
		})

		if err != nil {
			log.Fatal(err)
		}

		c.Start()

		select {}*/
}
