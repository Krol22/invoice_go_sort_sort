package llm

import (
	"fmt"

	"github.com/krol22/invoice_go_sort_sort/ai"
)

type AnalyzeInvoiceLLMInput struct {
  Invoice string
}

type AnalyzeInvoiceLLMOutputSchema struct {
  InvoiceDate OutputSchemaField `json:"invoiceDate"`
}

type AnalyzeInvoiceLLMOutputResponse struct {
  InvoiceDate string `json:"invoiceDate"`
}

type AnalyzeInvoiceLLM struct {
  BaseLLM
}

func (b *AnalyzeInvoiceLLM) GenerateChat() ([]ai.Message, error) {
  input, ok := b.inputData.(*AnalyzeInvoiceLLMInput)

  if !ok {
    return nil, fmt.Errorf("invalid input data")
  }

  messages := []ai.Message{
    {
      Role:    "user",
      Content: `
        You're a specialist in analysing the invoices and you're given a task of extracing the creation date of the invoice.  
      `,
    },
    {
      Role:    "user",
      Content: `
        Analyze the following invoice:
        <invoice>
        ` + 
        input.Invoice + `
        </invoice>

        Return the date in the following format 'YYYY-MM-DD'
        `,
    },
  }

  return messages, nil
}

func (b *AnalyzeInvoiceLLM) GetOutputSchema() map[string]interface{} {
  return map[string]interface{}{
    "date": map[string]interface{} {
      "type": "string",
      "description": "The creation date of the invoice",
    },
  }
}

func (b *AnalyzeInvoiceLLM) GetOutput() *AnalyzeInvoiceLLMOutputResponse {
  return &AnalyzeInvoiceLLMOutputResponse{
    InvoiceDate: b.aiResponse.JsonOutput["date"].(string),
  }
}

func NewAnalyzeInvoiceLLM(inputData interface{}) *AnalyzeInvoiceLLM {
  return &AnalyzeInvoiceLLM{
    BaseLLM: BaseLLM{
      inputData: inputData,
    },
  }
}
