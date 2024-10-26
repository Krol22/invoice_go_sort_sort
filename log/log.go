package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/krol22/automate_firma/notifications"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

var once sync.Once
var log zerolog.Logger

type FailureHook struct{}

func (f *FailureHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level == zerolog.FatalLevel && msg != "" {
		notifications.SendAlert("AutomateFirma script failed!")
	}
}

func Get() zerolog.Logger {
	once.Do(func() {
		var output io.Writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}

		if os.Getenv("ENV") != "development" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				fmt.Println("Error getting home directory:", err)
				return
			}

			logPath := filepath.Join(homeDir, "Library/Logs/com.krol22.automate_firma")
			os.MkdirAll(logPath, 0755)

			fileLogger := &lumberjack.Logger{
				Filename:   logPath + "/prod.log",
				MaxSize:    5, // megabytes
				MaxBackups: 10,
				MaxAge:     30,
				Compress:   true,
			}

			output = zerolog.MultiLevelWriter(os.Stderr, fileLogger)
		}

		log = zerolog.New(os.Stdout).With().Timestamp().Logger().Output(output).Hook(&FailureHook{})
	})

	return log
}
