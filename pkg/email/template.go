package email

import (
	"bytes"
	_ "embed"
	"html/template"
)

//go:embed templates/otp_email.html
var otpEmailHTML string

type otpEmailData struct {
	Title   string
	Message string
	Code    string
	Year    int
}

func RenderOTPEmail(title, message, code string, year int) (string, error) {

	tmpl, err := template.New("otp").Parse(otpEmailHTML)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, otpEmailData{Title: title, Message: message, Code: code, Year: year}); err != nil {
		return "", err
	}

	return buf.String(), nil
}
