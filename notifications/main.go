package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func SendAlert(message string) error {
  user_key := os.Getenv("PUSHOVER_USER_KEY")
  api_token := os.Getenv("PUSHOVER_API_TOKEN")

  body := map[string]string{
    "token": api_token,
    "user": user_key,
    "message": message,
  }

  jsonBody, err := json.Marshal(body)
  if err != nil {
    return err
  }
  

  fmt.Println("Sending alert: ", message, body)
  url := "https://api.pushover.net/1/messages.json"
  req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
  if err != nil {
    return err
  }

  req.Header.Set("Content-Type", "application/json")

  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    fmt.Println("Error sending alert: ", err)
    return err
  }

  respBody, err := io.ReadAll(resp.Body)
  fmt.Println("Response: ", string(respBody))
  defer resp.Body.Close()


  return nil
}
