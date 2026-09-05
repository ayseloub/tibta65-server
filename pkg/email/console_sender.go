package email

import (
	"context"
	"fmt"
)

type ConsoleEmailSender struct{}

func NewConsoleEmailSender() *ConsoleEmailSender {
	return &ConsoleEmailSender{}
}

func (s *ConsoleEmailSender) Send(ctx context.Context, to, subject, body string) error {
	fmt.Println("========== EMAIL (DEV MODE, gak beneran dikirim) ==========")
	fmt.Printf("To: %s\nSubject: %s\n\n%s\n", to, subject, body)
	fmt.Println("=============================================================")
	return nil
}
