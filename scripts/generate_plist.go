package main

import (
	"os"
	"path/filepath"

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
	StartCalendarInterval []CalendarInterval `plist:"StartCalendarInterval"`
}

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	agent := LaunchAgent{
		Label: "com.krol22.invoice_go_sort_sort",
		ProgramArguments: []string{
			filepath.Join(homeDir, ".scripts/invoice_go_sort_sort/invoice_go_sort_sort"),
		},
		RunAtLoad: true,
		StartCalendarInterval: []CalendarInterval{
			{Hour: 0, Minute: 0},
			{Hour: 6, Minute: 0},
			{Hour: 12, Minute: 0},
			{Hour: 18, Minute: 0},
		},
	}

	file, err := os.Create("dist/com.krol22.invoice_go_sort_sort.plist")
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
