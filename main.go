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
	"github.com/krol22/invoice_go_sort_sort/ai"
	"github.com/krol22/invoice_go_sort_sort/ai/llm"
	"github.com/krol22/invoice_go_sort_sort/email"
	"github.com/krol22/invoice_go_sort_sort/env"
	"github.com/krol22/invoice_go_sort_sort/log"
	"github.com/krol22/invoice_go_sort_sort/state"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
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
	icloudPath := env.Get("ICLOUD_PATH")

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


func upgradePDFVersion() ([]byte, error) {
		tmpDir := os.TempDir()
		pdfPath := tmpDir + "/pdf.pdf"
		inputFile, err := os.Open(pdfPath)
    if err != nil {
        return nil, fmt.Errorf("error opening input file: %v", err)
    }
    defer inputFile.Close()

    // Create configuration with version 1.4
    conf := model.NewDefaultConfiguration()
    conf.Version = "1.4"

    // Get page count of input file
    pageCount, err := api.PageCount(inputFile, conf)
    if err != nil {
        return nil, fmt.Errorf("error getting page count: %v", err)
    }

    // Create a list of pages to copy (all pages)
    pages := make([]string, pageCount)
    for i := 0; i < pageCount; i++ {
        pages[i] = fmt.Sprintf("%d", i+1)
    }

		tempDir, err := os.MkdirTemp("", "pdf_upgrade_*")
    if err != nil {
        return nil, fmt.Errorf("error creating temp directory: %v", err)
    }
    // defer os.RemoveAll(tempDir) // Clean up when done

		outputFilename := "upgraded.pdf"

    // Extract pages to temp directory
    err = api.ExtractPages(inputFile, tempDir, outputFilename, pages, conf)
    if err != nil {
        return nil, fmt.Errorf("error extracting pages: %v", err)
    }

		// Get list of extracted files
    var inFiles []string
    for i := 1; i <= pageCount; i++ {
        inFiles = append(inFiles, filepath.Join(tempDir, fmt.Sprintf("upgraded_page_%d.pdf", i)))
    }

    // Create buffer for final output
    var buf bytes.Buffer

    // Merge all pages into a single PDF
    err = api.Merge("", inFiles, &buf, conf, false)
    if err != nil {
        return nil, fmt.Errorf("error merging pages: %v", err)
    }

    fmt.Printf("Successfully created new PDF version 1.4 with %d pages\n", pageCount)
    return buf.Bytes(), nil
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
	emailManager, err := email.NewEmailManager(env.Get("EMAIL"), env.Get("API_KEY"))	
	if err != nil {
		return nil, fmt.Errorf("failed to create email manager: %v", err)
	}
	defer emailManager.Logout()

	messages, err := emailManager.GetFilteredMessages(
		env.Get("FORWARDED_FROM_EMAIL"),
		lastRun.AddDate(0, 0, -1),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %v", err)
	}

	return messages, nil
}

func analyzeAttachment(pdfText string, attachment *email.Attachment) error {
	anthropicClient := ai.NewClient(env.Get("ANTHROPIC_KEY"))
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

	l.Print("Starting invoice sorting...")

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
					l.Print("Upgrading PDF version...")
					v14, err := upgradePDFVersion()
					if err != nil {
						l.Fatal().Err(err).Msg("Error upgrading PDF version")
					}

					pdfText, err = extractTextFromPDF(v14)
					if err != nil {
						l.Fatal().Err(err).Msg("Error extracting text from PDF")
					}
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
