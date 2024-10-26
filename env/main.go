package env

import (
	"os"
	"sync"
)

var (
    ForwardedFromEmail string
    ForwardedToEmail   string
    Email             string
    IcloudPath        string
    ApiKey            string
    AnthropicKey      string
    AnthropicVersion  string
    PushoverApiToken  string
    PushoverUserKey   string
)

var once sync.Once

func Get(key string) string {
  if os.Getenv("ENV") == "development" {
    return os.Getenv(key)
  }

  switch key {
  case "FORWARDED_FROM_EMAIL":
    return ForwardedFromEmail
  case "FORWARDED_TO_EMAIL":
    return ForwardedToEmail
  case "EMAIL":
    return Email
  case "ICLOUD_PATH":
    return IcloudPath
  case "API_KEY":
    return ApiKey
  case "ANTHROPIC_KEY":
    return AnthropicKey
  case "ANTHROPIC_VERSION":
    return AnthropicVersion
  case "PUSHOVER_API_TOKEN":
    return PushoverApiToken
  case "PUSHOVER_USER_KEY":
    return PushoverUserKey
  default:
    return ""
  }
}
