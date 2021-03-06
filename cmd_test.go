package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
)

func init() {
	Logger.Formatter = new(logrus.TextFormatter)
	Logger.Level = logrus.Debug
}

func TestMonitorCmd(t *testing.T) {
	rc := NewRedisCmds()
	stopper := make(chan bool, 1)
	replies := rc.MonitorCmd(stopper)
	timeout := time.AfterFunc(1*time.Second, func() {
		stopper <- true
		t.Fatalf("Failed timout execed reply")
	})
	// Push a bunch of calls for monitor
	for i := 0; i < 100; i++ {
		rc.InfoCmd()
	}
	for r := range replies {
		f := strings.Index(r, "INFO")
		if f > -1 {
			break
		}
	}
	stopper <- true
	timeout.Stop()
}

func TestInfoCmd(t *testing.T) {
	rc := NewRedisCmds()
	s, err := rc.InfoCmd()
	if err != nil {
		t.Fatalf("Error running InfoCmd %s", err)
	}
	if strings.Index(s, "redis_version") < 0 {
		t.Fatalf("Expected to find 'redis_version' in the InfoCmd")
	}
}

func TestSlowlogCmd(t *testing.T) {
	rc := NewRedisCmds()
	genSize := 100
	for x := 0; x < genSize; x++ {
		k := fmt.Sprintf("_slowlog_test_%d", x)
		rc.conn().Do("SET", k, x)
		rc.conn().Do("GET", k, x)
	}
	rc.conn().Flush()
	for x := 0; x < genSize; x++ {
		go func() {
			rc.conn().Send("KEYS", "*slowlog*")
		}()
	}
	for x := 0; x < genSize; x++ {
		k := fmt.Sprintf("_slowlog_test_%d", x)
		rc.conn().Send("DEL", k)
	}

	logs, err := rc.SlowlogCmd()
	if err != nil {
		t.Fatalf("Failed to call slowlogcmd error: %s", err)
	}
	if logs == nil {
		t.Fatalf("No slowlogs :(")
	}
	if len(logs) < 1 {
		t.Errorf("Expected 10 logs got %d", len(logs))
	}

}
