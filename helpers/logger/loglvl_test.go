package logger

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestLogLevelDefaultInfo(t *testing.T) {
	os.Unsetenv("LOG_LEVEL")
	if logLevel() != logrus.InfoLevel {
		t.Fatalf("default level = %v, want info", logLevel())
	}
	os.Setenv("LOG_LEVEL", "debug")
	if logLevel() != logrus.DebugLevel {
		t.Fatalf("LOG_LEVEL=debug → %v, want debug", logLevel())
	}
	os.Setenv("LOG_LEVEL", "bogus")
	if logLevel() != logrus.InfoLevel {
		t.Fatalf("bogus → %v, want info fallback", logLevel())
	}
	os.Unsetenv("LOG_LEVEL")
}

func TestLogMaxAgeDefault3Days(t *testing.T) {
	os.Unsetenv("LOG_MAX_AGE_DAYS")
	if logMaxAge() != time.Hour*24*3 {
		t.Fatalf("default = %v, want 72h", logMaxAge())
	}
	os.Setenv("LOG_MAX_AGE_DAYS", "10")
	if logMaxAge() != time.Hour*24*10 {
		t.Fatalf("10 days → %v", logMaxAge())
	}
	os.Setenv("LOG_MAX_AGE_DAYS", "0")
	if logMaxAge() != time.Hour*24*3 {
		t.Fatalf("0 → %v, want 72h fallback", logMaxAge())
	}
	os.Unsetenv("LOG_MAX_AGE_DAYS")
}
