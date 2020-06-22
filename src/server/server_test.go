package server

import (
	"fmt"
	"testing"

	"github.com/robfig/cron/v3"
)

func TestCron(T *testing.T) {
	c := cron.New(cron.WithSeconds())

	c.AddFunc("*/1 * * * * *", func() {
		fmt.Println(1)
		fmt.Println(c.Entries())
	})
	c.AddFunc("*/5 * * * * *", func() {
		fmt.Println(5)
		fmt.Println(c.Entries())
	})
	c.AddFunc("*/10 * * * * *", func() {
		fmt.Println(10)
		fmt.Println(c.Entries())
	})
	c.Start()
	defer c.Stop()
	for {
	}
}