package logger

import (
	"fmt"
)

// LogAndPrint logs to structured logger AND prints to terminal (stdout).
func LogAndPrint(level string, msg string, err error) {
	icon := ""
	switch level {
	case "error":
		icon = "🚨"
	case "warn":
		icon = "⚠️"
	case "info":
		icon = "ℹ️"
	case "debug":
		icon = "🐛"
	default:
		icon = "🔸"
	}

	if err != nil {
		fmt.Printf("%s %s: %v\n", icon, msg, err)
	} else {
		fmt.Printf("%s %s\n", icon, msg)
	}

	switch level {
	case "error":
		Error(msg, err)
	case "warn":
		Warn(msg, err)
	case "info":
		Info(msg)
	case "debug":
		Debug(msg)
	default:
		Info(msg)
	}
}

func PrintError(msg string, err error) error {
	LogAndPrint("error", msg, err)
	return err
}
