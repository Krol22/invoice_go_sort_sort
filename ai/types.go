package ai

type LLM interface {
  GenerateChat() ([]Message, error)
  GetOutputSchema() map[string]interface{}
  GetMaxTokens() int

  SetAiResponse(aiResponse *AiResponse)
  GetAiResponse() *AiResponse
}
