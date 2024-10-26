package state

import (
	"os/exec"
	"strings"
	"time"
)

func SaveLastRun() error {
  cmd := exec.Command("defaults", "write", "com.krol22.invoice_go_sort_sort", "last_run", time.Now().Format("2006-01-02"))
  return cmd.Run()
}

func LoadLastRun() (time.Time, error) {
  cmd := exec.Command("defaults", "read", "com.krol22.invoice_go_sort_sort", "last_run")
  out, err := cmd.Output()

  if err != nil {
    return time.Now(), nil
  }

  return time.Parse(time.DateOnly, strings.TrimSpace(string((out))))
}
