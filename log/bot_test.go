package log

import (
	"testing"
	"time"
)

func TestSendToBot(t *testing.T) {
	Error("32322")
	time.Sleep(time.Second * 3)
}
