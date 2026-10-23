package report

import (
	"fmt"
	"strings"
	"time"

	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/testing-microservice-soap/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	gomail "gopkg.in/mail.v2"
)

func mapToString(m map[string]string) string {
	parts := make([]string, 0, len(m))
	parts = append(parts, "Report "+time.Now().Format("2006-01-02 15:04:05")+":	")
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, " ")
}

func SendReportTG(report map[string]string) {
	cfg := config.GetTgconfig()
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		logger.LogDebug("Not found tg bot", err)
		return
	}

	users := []int64{}
	msgText := mapToString(report)

	for _, userID := range users {
		msg := tgbotapi.NewMessage(userID, msgText)
		_, err := bot.Send(msg)
		if err != nil {
			logger.LogDebug("Not sended to tg bot", err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	logger.LogDebug("Report sended successfully!")
}

func SendReportMail(report map[string]string) {

	msgText := mapToString(report)

	cfg := config.GetMailconfig()

	m := gomail.NewMessage()

	m.SetHeader("From", cfg.From)
	m.SetHeader("To", cfg.To)
	m.SetHeader("Subject", "Test report! "+time.Now().String())

	m.SetBody("text/plain", msgText)
	logger.LogDebug("msg", m)

	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.From, cfg.Pass)

	if err := dialer.DialAndSend(m); err != nil {
		panic(err)
	}

	logger.LogDebug("Report sended successfully!")
}
