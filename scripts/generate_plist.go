package main

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"howett.net/plist"
)

type CalendarInterval struct {
	Hour   int `plist:"Hour"`
	Minute int `plist:"Minute"`
}

type LaunchAgent struct {
	Label                 string             `plist:"Label"`
	ProgramArguments      []string           `plist:"ProgramArguments"`
	RunAtLoad             bool               `plist:"RunAtLoad"`
	EnvironmentVariables  map[string]string  `plist:"EnvironmentVariables,omitempty"`
	StartCalendarInterval []CalendarInterval `plist:"StartCalendarInterval"`
}

func main() {
	env, err := godotenv.Read()
	if err != nil {
		panic(err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	agent := LaunchAgent{
		Label: "com.krol22.automate_firma",
		ProgramArguments: []string{
			filepath.Join(homeDir, ".scripts/automate_firma/automate_firma"),
		},
		RunAtLoad: true,
		EnvironmentVariables: map[string]string{
			"FORWARDED_FROM_EMAIL": env["FORWARDED_FROM_EMAIL"],
			"FORWARDED_TO_EMAIL":   env["FORWARDED_TO_EMAIL"],
			"EMAIL":                env["EMAIL"],
			"ICLOUD_PATH":          env["ICLOUD_PATH"],
			"API_KEY":              env["API_KEY"],
			"ANTHROPIC_KEY":        env["ANTHROPIC_KEY"],
			"PUSHOVER_USER_KEY":    env["PUSHOVER_USER_KEY"],
			"PUSHOVER_API_TOKEN":   env["PUSHOVER_API_TOKEN"],
		},
		StartCalendarInterval: []CalendarInterval{
			{Hour: 0, Minute: 0},
			{Hour: 6, Minute: 0},
			{Hour: 12, Minute: 0},
			{Hour: 18, Minute: 0},
		},
	}

	file, err := os.Create("dist/com.krol22.automate_firma.plist")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	encoder := plist.NewEncoder(file)
	encoder.Indent("    ")
	if err := encoder.Encode(agent); err != nil {
		panic(err)
	}
}
