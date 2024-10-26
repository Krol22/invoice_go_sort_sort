package llm

import "github.com/krol22/automate_firma/ai"

type BaseLLM struct {
  inputData interface{}
  aiResponse *ai.AiResponse
}

func (b *BaseLLM) GetMaxTokens() int {
  return 4096
}

func (b *BaseLLM) SetAiResponse(aiResponse *ai.AiResponse) {
  b.aiResponse = aiResponse
}

func (b *BaseLLM) GetAiResponse() *ai.AiResponse {
  return b.aiResponse
}

type OutputSchemaField struct {
  Type string `json:"type"`
  Description string `json:"description"`
}
