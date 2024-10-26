package email

import (
	"fmt"
	"io"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/krol22/invoice_go_sort_sort/log"
)

var l = log.Get()

type EmailMessage struct {
	Message *imap.Message
	Attachments []Attachment
}

type Attachment struct {
	Filename string
	Content []byte
}

type EmailManager struct {
  client *client.Client
}

func NewEmailManager(email, password string) (*EmailManager, error) {
  em := &EmailManager{}
  if err := em.Login(email, password); err != nil {
    return nil, err
  }
  return em, nil
}

func (e *EmailManager) Login(email, password string) (error) {
	l.Print("Connecting to imap.gmail.com server.")
	c, err := client.DialTLS("imap.gmail.com:993", nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
  e.client = c

	l.Print("Logging into ", email, " account.")
	if err := c.Login(email, password); err != nil {
		return fmt.Errorf("failed to login: %v", err)
	}

  return nil
}

func (e *EmailManager) Logout() error {
	l.Print("Logging out from email.")
  if err := e.client.Logout(); err != nil {
    return fmt.Errorf("failed to logout: %v", err)
  }
  return nil
}

func (e *EmailManager) GetFilteredMessages(
	email string,
	to string,
	dateFrom time.Time,
) ([]*EmailMessage, error) {
	l.Print("Opening email inbox.")
	_, err := e.client.Select("INBOX", false)
	if err != nil {
		return nil, fmt.Errorf("failed to select INBOX: %v", err)
	}

	// Set search criteria
	criteria := imap.NewSearchCriteria()
	criteria.Since = dateFrom
	criteria.Header.Set("From", email)
	criteria.Header.Set("To", to)

	l.Print("Searching for messages using criteria:", map[string]string{
		"From":  email,
		"To":  to,
		"Since": dateFrom.Format("2006-01-02"),
	})
	uids, err := e.client.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %v", err)
	}

	messageWord := "messages"
	if len(uids) == 1 {
		messageWord = "message"
	}

	l.Print("Found ", len(uids), " ", messageWord, ".")
	if len(uids) == 0 {
		return []*EmailMessage{}, nil
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uids...)

	l.Print("Fetching messages.")
	messages := make(chan *imap.Message, len(uids))
	done := make(chan error, 1)
	go func() {
		done <- e.client.Fetch(seqSet, []imap.FetchItem{imap.FetchEnvelope, imap.FetchRFC822}, messages)
	}()

	var result []*EmailMessage
	for msg := range messages {
		emailMsg, err := e.processMessage(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to process message: %v", err)
		}
		result = append(result, emailMsg)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %v", err)
	}

	return result, nil
}

func (e *EmailManager) processMessage(msg *imap.Message) (*EmailMessage, error) {
	emailMsg := &EmailMessage{
		Message: msg,
	}

	l.Print("Processing message ", msg.Uid, " from ", msg.Envelope.From[0].Address(), " with subject: '", msg.Envelope.Subject, "'")
	raw := msg.GetBody(&imap.BodySectionName{})
	if raw == nil {
			return nil, fmt.Errorf("failed to get raw message body")
	}

	for _, part := range msg.Body {
		mr, err := mail.CreateReader(part)
		if err != nil {
			return nil, fmt.Errorf("failed to create mail reader: %v", err)
		}

		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("failed to get next part: %v", err)
			}

			switch h := p.Header.(type) {
			case *mail.AttachmentHeader:
				// This is an attachment
				filename, _ := h.Filename()
				content, err := io.ReadAll(p.Body)
				l.Print("Found attachment: ", filename, " with size: ", len(content))
				if err != nil {
					return nil, fmt.Errorf("failed to read attachment content: %v", err)
				}
				attachment := Attachment{
					Filename: filename,
					Content:  content,
				}
				emailMsg.Attachments = append(emailMsg.Attachments, attachment)
			}
		}
	}

	return emailMsg, nil
}
