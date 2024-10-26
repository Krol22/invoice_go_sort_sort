package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/krol22/automate_firma/ai"
	"github.com/krol22/automate_firma/ai/llm"
	"github.com/krol22/automate_firma/email"
	"github.com/krol22/automate_firma/log"
	"github.com/krol22/automate_firma/state"
)

var l = log.Get()

var polishMonths = map[time.Month]string{
	time.January:   "styczeń",
	time.February:  "luty",
	time.March:     "marzec",
	time.April:     "kwiecień",
	time.May:       "maj",
	time.June:      "czerwiec",
	time.July:      "lipiec",
	time.August:    "sierpień",
	time.September: "wrzesień",
	time.October:   "październik",
	time.November:  "listopad",
	time.December:  "grudzień",
}

func getInvoiceMonthPath(invoiceDate string) string {
	icloudPath := os.Getenv("ICLOUD_PATH")

	date, err := time.Parse("2006-01-02", invoiceDate)
	if err != nil {
		l.Fatal().Err(err).Msg("Error parsing date")
	}

	year := date.Year()
	month := date.Month()

	monthName := polishMonths[month]

	return icloudPath + "/Documents/Firma/" + fmt.Sprint(year) + "/dokumenty_" + monthName
}

func createFoldersIfNecessary(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0755)
	}
}

func extractTextFromPDF(file []byte) (string, error) {
	tmpDir := os.TempDir()
	pdfPath := tmpDir + "/pdf.pdf"
	err := os.WriteFile(pdfPath, file, 0644)

	if err != nil {
		return "", fmt.Errorf("error writing pdf file: %v", err)
	}

	l.Print("Executing pdf2txt.py")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %v", err)
	}
	cmd := exec.Command(filepath.Join(homeDir, "/Library/Python/3.9/bin/pdf2txt.py"), pdfPath)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error executing pdf2txt.py: %v\nStderr: %s", err, stderr.String())
	}

	l.Print("Extracted content from the PDF: ", len(stdout.String()), " characters.")
	return stdout.String(), nil
}

func getEmailInvoices(lastRun time.Time) ([]*email.EmailMessage, error) {
	emailManager, err := email.NewEmailManager(os.Getenv("EMAIL"), os.Getenv("API_KEY"))	
	if err != nil {
		return nil, fmt.Errorf("failed to create email manager: %v", err)
	}
	defer emailManager.Logout()

	messages, err := emailManager.GetFilteredMessages(
		os.Getenv("FORWARDED_FROM_EMAIL"),
		os.Getenv("FORWARDED_TO_EMAIL"),
		lastRun.AddDate(0, 0, -1),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %v", err)
	}

	return messages, nil
}

func analyzeAttachment(pdfText string, attachment *email.Attachment) error {
	anthropicClient := ai.NewClient(os.Getenv("ANTHROPIC_KEY"))
	analyzeInvoice := llm.NewAnalyzeInvoiceLLM(&llm.AnalyzeInvoiceLLMInput{
		Invoice: pdfText,
	})

	_, err := anthropicClient.RunLLM(analyzeInvoice)

	if err != nil {
		l.Fatal().Err(err).Msg("Failed to analyze invoice")
	}

	outputData := analyzeInvoice.GetOutput()

	invoicePath := getInvoiceMonthPath(outputData.InvoiceDate)
	l.Print("Selecting path for the invoice (", attachment.Filename, "): ", invoicePath)
	createFoldersIfNecessary(invoicePath)

	l.Print("Saving invoice to: ", invoicePath+"/"+attachment.Filename)
	err = os.WriteFile(invoicePath+"/"+attachment.Filename, attachment.Content, 0644)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to save invoice")
	}

	return nil
}

func main() {
	for range 10 {
		l.Print("#################################")
	}

	l.Print("Starting automate firma script...")


	if os.Getenv("ENV") == "development" {
		l.Print("Loading .env file...")
		err := godotenv.Load()
		if err != nil {
			l.Fatal().Err(err).Msg("Error loading .env file")
		}
	}

	lastRun, _ := state.LoadLastRun()

	l.Print("Starting fetching email invoices.")
	emailMessages, err := getEmailInvoices(lastRun)
	if err != nil {
		l.Fatal().Err(err).Msg("Error fetching email invoices")
	}

	for _, emailMessage := range emailMessages {
		for _, attachment := range emailMessage.Attachments {
			if strings.HasSuffix(strings.ToLower(attachment.Filename), ".pdf") {
				l.Print("Processing PDF attachment: ", attachment.Filename)
				pdfText, err := extractTextFromPDF(attachment.Content)

				if err != nil {
					l.Fatal().Err(err).Msg("Error extracting text from PDF")
				}

				err = analyzeAttachment(pdfText, &attachment)

				if err != nil {
					l.Fatal().Err(err).Msg("Error analyzing attachment")
				}
			}
		}
	}

	err = state.SaveLastRun()
	if err != nil {
		l.Fatal().Err(err).Msg("Error saving last run")
	}

	l.Print("Finished!")
}
