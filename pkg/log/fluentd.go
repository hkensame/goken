package log

import (
	"fmt"
	"log"

	"github.com/fluent/fluent-logger-golang/fluent"
)

func t() {
	fluentd, err := fluent.New(fluent.Config{
		FluentHost: "192.168.199.128",
		FluentPort: 24224,
	})
	if err != nil {
		panic(err)
	}
	defer fluentd.Close()

	// 记录日志
	logData := map[string]interface{}{
		"level":   "info",
		"message": "This is a test log",
	}

	// 发送日志到 Fluentd
	err = fluentd.Post("myapp.logs", logData)
	if err != nil {
		log.Fatal("Error sending log to Fluentd:", err)
	}

	// 输出日志，确认日志已发送
	fmt.Println("Log sent to Fluentd!")
}
