package logger

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{Debug, "DEBUG "},
		{Info, "INFO "},
		{Warning, "WARNING "},
		{Error, "ERROR "},
	}

	t.Parallel()
	for _, test := range tests {
		test := test
		t.Run(string(test.level), func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			logger := New(string(test.level), &buf)

			// Log a message
			logger.Log("Test message")

			// Check if the logged message contains the expected level prefix
			if !strings.HasPrefix(buf.String(), test.expected) {
				t.Errorf("Expected message to start with '%s', but got: '%s'", test.expected, buf.String())
			}
		})
	}
}

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name                string
		level               string
		funcName            string
		message             string
		expectedMsgContains string
	}{
		{
			name:                "error msg",
			level:               "ERROR",
			funcName:            "Error",
			message:             "Test message", // Log a message
			expectedMsgContains: "ERROR [",      // ERROR [2024-01-18 07:36:31] Test message\n
		},
		{
			name:                "skipp_warning",
			level:               "ERROR",
			funcName:            "Warning",
			message:             "Test message",
			expectedMsgContains: "",
		},
		{
			name:                "skipp_info",
			level:               "ERROR",
			funcName:            "Info",
			message:             "Test message",
			expectedMsgContains: "",
		},
		{
			name:                "debug",
			level:               "DEBUG",
			funcName:            "Debug",
			message:             "Test message",
			expectedMsgContains: "DEBUG [", // DEBUG [2024-01-18 07:37:33] Test message\n
		},
		{
			name:                "skipp_debug",
			level:               "ERROR",
			funcName:            "Debug",
			message:             "Test message",
			expectedMsgContains: "",
		},
		{
			name:                "skipp_debug2",
			level:               "INFO",
			funcName:            "Debug",
			message:             "Test message",
			expectedMsgContains: "",
		},
		{
			name:                "info",
			level:               "INFO",
			funcName:            "Info",
			message:             "Test message",
			expectedMsgContains: "INFO [", // INFO [2024-01-18 07:38:13] Test message\n
		},
		{
			name:                "skipp_info2",
			level:               "WARNING",
			funcName:            "Info",
			message:             "Test message",
			expectedMsgContains: "",
		},
		{
			name:                "warning",
			level:               "WARNING",
			funcName:            "Warning",
			message:             "Test message",
			expectedMsgContains: "WARNING [", // WARNING [2024-01-18 07:38:52] Test message\n
		},
	}

	t.Parallel()
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("case %s", test.name), func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			logger := New(test.level, &buf)
			switch test.funcName {
			case "Error":
				logger.Error(test.message)
			case "Warning":
				logger.Warning(test.message)
			case "Info":
				logger.Info(test.message)
			case "Debug":
				logger.Debug(test.message)
			}

			// Check if the logged message contains the expected level prefix
			require.Contains(t, buf.String(), test.expectedMsgContains)
		})
	}
}

func TestFatalf(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		log.Fatalf("invalid log level")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestFatalf") //nolint:gosec
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	var e *exec.ExitError
	if errors.As(err, &e) && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}
