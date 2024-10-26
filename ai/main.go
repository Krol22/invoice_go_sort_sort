package ai

type Message struct {
  Role string `json:"role"`
  Content string `json:"content"`
}

type AiResponse struct {
  Message Message
  JsonOutput map[string]interface{}
}

