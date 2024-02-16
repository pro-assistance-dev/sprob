package email

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"strings"

	"github.com/pro-assistance/pro-assister/config"
)

type AuthMethod string

const (
	PlainAuthMethod AuthMethod = "PlainAuth"
	LoginAuthMethod AuthMethod = "LoginAuth"
)

// Email struct
// https://medium.com/@dhanushgopinath/sending-html-emails-using-templates-in-golang-9e953ca32f3d
type Email struct {
	config  config.Email
	request request
}

func NewEmail(c config.Email) *Email {
	return &Email{config: c}
}

// request struct
type request struct {
	From    string
	To      []string
	Subject string
	Body    string
}

// SetRequest struct
func (e *Email) SendEmail(to []string, subject string, body string) error {
	e.request = request{To: to, Subject: subject, Body: body}
	return e.sendEmail()
}

// SendEmail func
func (e *Email) sendEmail() error {
	auth := smtp.PlainAuth(
		"",
		e.config.From,
		e.config.Password,
		e.config.Server,
	)
	if e.config.AuthMethod == string(LoginAuthMethod) {
		auth = LoginAuth(e.config.From, e.config.Password)
	}
	header := map[string]string{}
	header["To"] = strings.Join(e.request.To, ",")
	header["From"] = e.config.From
	header["Subject"] = e.request.Subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=\"utf-8\""
	// header["Content-Transfer-Encoding"] = "base64"
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	// message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(e.request.Body))
	message += "\r\n" + e.request.Body
	servername := fmt.Sprintf("%s:%s", e.config.Server, e.config.Port)
	host, _, _ := net.SplitHostPort(servername)

	tlsconfig := &tls.Config{ //nolint:gosec
		InsecureSkipVerify: false, //nolint:gosec
		ServerName:         host,
	}
	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		return err
	}
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	// Auth
	if err = c.Auth(auth); err != nil {
		return err
	}
	// To && From
	if err = c.Mail(e.config.From); err != nil {
		return err
	}
	for _, t := range e.request.To {
		if err = c.Rcpt(t); err != nil {
			return err
		}
	}
	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	err = c.Quit()
	if err != nil {
		return err
	}

	if e.config.WriteTestFile {
		err = os.WriteFile("./application-generate_send.html", []byte(message), 0600)
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}
