package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/krol22/automate_firma/log"
	"github.com/krol22/automate_firma/utils"
)

type AnthropicClient struct {
  baseUrl string
  apiKey string
}

type anthropicResponse struct {
  Id string `json:"id"`
  Type string `json:"type"`
  Role string `json:"role"`
  Model string `json:"model"`
  Content []anthropicContent `json:"content"`
  StopReason string `json:"stop_reason"`
  StopSequence string `json:"stop_sequence"`
  Usage anthropicUsage `json:"usage"`
  ToolUse interface{} `json:"tool_use"`
}

type anthropicUsage struct {
  InputTokens int `json:"input_tokens"`
  OutputTokens int `json:"output_tokens"`
}

type anthropicContent struct {
  Type string `json:"type"` 
  Text string `json:"text"`
  Id string `json:"id"`
  Name string `json:"name"`
  Input map[string]interface{} `json:"input"`
}

type anthropicTool struct {
  Name string `json:"name"`
  Description string `json:"description"`
  InputSchema inputSchema `json:"input_schema"`
}

type inputSchema struct {
  Type string `json:"type"`
  Properties map[string]inputSchemaProperty `json:"properties"`
  Required []string `json:"required"`
}

type inputSchemaProperty struct {
  Type string `json:"type"`
  Description string `json:"description"`
}

const MODEL = "claude-3-5-sonnet-20240620"

var l = log.Get()

func NewClient(apiKey string) *AnthropicClient {
  return &AnthropicClient{
    baseUrl: "https://api.anthropic.com",
    apiKey: apiKey,
  }
}

func (c *AnthropicClient) createRequest(method string, body map[string]interface{}) (*http.Request, error) {
  jsonBody, err := json.Marshal(body)
  if err != nil {
    return nil, err
  }

  url := c.baseUrl + "/v1/messages/"
  req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
  if err != nil {
    return nil, err
  }

  l.Print("Sending request to: ", url)
  l.Print("With body: ", utils.PrettyPrint(body))

  req.Header.Set("x-api-key", c.apiKey)
  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("anthropic-version", "2023-06-01")

  return req, nil
}

func (c *AnthropicClient) mapResponse(response anthropicResponse) (*AiResponse, error) {
  var res *AiResponse

  l.Print("Got response: ", utils.PrettyPrint(response))

  switch response.Content[0].Type {
    case "text":
      res = &AiResponse {
        Message: Message {
          Role: response.Role,
          Content: response.Content[0].Text,
        },
      }
    case "tool_use":
      res = &AiResponse {
        JsonOutput: response.Content[0].Input,
      }
  }

  return res, nil 
}

func (c *AnthropicClient) sendRequest(req *http.Request) (*AiResponse, error) {
  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    return nil, err
  }
  defer resp.Body.Close()

  // Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Print the response
  if resp.StatusCode != 200 {
    return nil, fmt.Errorf("request failed with status code: %d\n%s", resp.StatusCode, string(body))
  }

  aResp := &anthropicResponse{}
  err = json.Unmarshal(body, aResp)
  if err != nil {
    return nil, fmt.Errorf("error unmarshalling response: %v", err)
  }

  return c.mapResponse(*aResp)
}

func (c *AnthropicClient) RunLLM(llm LLM) (*AiResponse, error) {
  chatMessages, err := llm.GenerateChat()

  if err != nil {
    return nil, err
  }

  requestBody := map[string]interface{}{
    "model": MODEL,
    "max_tokens": llm.GetMaxTokens(),
    "messages": chatMessages,
    "tool_choice": map[string]interface{} {
      "type": "tool",
      "name": "data_extractor",
      "disable_parallel_tool_use": false,
    },
  }

  outputSchema := llm.GetOutputSchema()
  if outputSchema != nil {
    inputSchemaProperties := make(map[string]inputSchemaProperty)
    for key, value := range outputSchema {
        if propMap, ok := value.(map[string]interface{}); ok {
            inputSchemaProperties[key] = inputSchemaProperty{
                Type:        propMap["type"].(string),
                Description: propMap["description"].(string),
            }
        }
    }
    
    // Prepare the input schema
    inputSchema := inputSchema{
        Type:       "object",
        Properties: inputSchemaProperties,
        Required:   make([]string, 0, len(inputSchemaProperties)),
    }
    
    // Add all properties to Required for this example
    for key := range inputSchemaProperties {
        inputSchema.Required = append(inputSchema.Required, key)
    }

    requestBody["tools"] = []anthropicTool{
      {
        Name: "data_extractor",
        Description: "extract the data to the exact provided format",
        InputSchema: inputSchema,
      },
    }
  }

  req, err := c.createRequest("POST", requestBody)
  if err != nil {
    return nil, err
  }

  aiResponse, err := c.sendRequest(req)

  if err != nil {
    return nil, err
  }

  llm.SetAiResponse(aiResponse)
  return aiResponse, nil
}

func (c *AnthropicClient) AskChat(messages []Message) (*AiResponse, error) {
  requestBody := map[string]interface{}{
    "model": MODEL,
    "max_tokens": 4096,
    "messages": messages,
  }

  req, err := c.createRequest("POST", requestBody)
  if err != nil {
    return nil, err
  }

  return c.sendRequest(req)
}
