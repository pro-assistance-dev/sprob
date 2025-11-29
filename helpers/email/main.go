package email

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"

	"github.com/pro-assistance-dev/sprob/config"
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
	From        string
	To          []string
	Subject     string
	Body        string
	Attachments map[string][]byte
}

// SetRequest struct
func (e *Email) SendEmail(to []string, subject string, body string) error {
	e.request = request{To: to, Subject: subject, Body: body}
	return e.sendEmail()
}

func (e *Email) SendEmailWithAttachments(to []string, subject string, body string, files []string) error {
	e.request = request{To: to, Subject: subject, Body: body}
	if len(files) > 0 {
		e.request.Attachments = make(map[string][]byte)
		for _, f := range files {
			err := e.request.AttachFile(f)
			if err != nil {
				return err
			}
		}
	}
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

	// Формируем правильные заголовки для HTML письма
	headers := make(map[string]string)
	headers["From"] = e.config.From
	headers["To"] = strings.Join(e.request.To, ", ")
	headers["Subject"] = e.request.Subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"utf-8\""
	headers["Content-Transfer-Encoding"] = "quoted-printable"

	// Собираем сообщение
	var message strings.Builder
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n") // разделитель заголовков и тела

	// Кодируем тело в quoted-printable для корректной передачи
	bodyEncoded := quotedprintable.NewWriter(&message)
	_, err := bodyEncoded.Write([]byte(e.request.Body))
	if err != nil {
		return err
	}
	bodyEncoded.Close()

	servername := fmt.Sprintf("%s:%s", e.config.Server, e.config.Port)
	host, _, _ := net.SplitHostPort(servername)

	tlsconfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	}
	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer c.Close()

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

	_, err = w.Write([]byte(message.String()))
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
		err = os.WriteFile("./application-generate_send.html", []byte(message.String()), 0o600)
		if err != nil {
			log.Printf("Error writing test file: %v", err)
		}
	}

	return nil
}

func (m *request) ToBytes(from string) []byte {
	buf := bytes.NewBuffer(nil)
	withAttachments := len(m.Attachments) > 0
	fmt.Fprintf(buf, "Subject: %s\n", m.Subject)
	fmt.Fprintf(buf, "To: %s\n", strings.Join(m.To, ","))
	fmt.Fprintf(buf, "From: %s\n", strings.Join([]string{from}, ","))
	// if len(m.CC) > 0 {
	// 	buf.WriteString(fmt.Sprintf("Cc: %s\n", strings.Join(m.CC, ",")))
	// }
	//
	// if len(m.BCC) > 0 {
	// 	buf.WriteString(fmt.Sprintf("Bcc: %s\n", strings.Join(m.BCC, ",")))
	// }
	//
	buf.WriteString("MIME-Version: 1.0\n")
	writer := multipart.NewWriter(buf)
	boundary := writer.Boundary()
	if withAttachments {
		fmt.Fprintf(buf, "Content-Type: multipart/mixed; boundary=%s\n", boundary)
		fmt.Fprintf(buf, "--%s\n", boundary)
	} else {
		buf.WriteString("Content-Type: text/plain; charset=utf-8\n")
	}

	fmt.Fprintf(buf, "\n\n--%s\n", boundary)
	fmt.Fprintf(buf, "Content-Type: text/html; charset=\"utf-8\" boundary=%s\n", boundary)
	buf.WriteString(m.Body)
	fmt.Fprintf(buf, "\n--%s", boundary)

	coder := base64.StdEncoding
	if withAttachments {
		for k, v := range m.Attachments {
			buf.WriteString("\r\n\r\n--" + boundary + "\r\n")

			ext := filepath.Ext(k)
			mimetype := mime.TypeByExtension(ext)
			if mimetype != "" {
				mime := fmt.Sprintf("Content-Type: %s\r\n", mimetype)
				buf.WriteString(mime)
			} else {
				buf.WriteString("Content-Type: application/octet-stream\r\n")
			}
			buf.WriteString("Content-Transfer-Encoding: base64\r\n")

			buf.WriteString("Content-Disposition: attachment; filename=\"=?UTF-8?B?")
			buf.WriteString(coder.EncodeToString([]byte(k)))
			buf.WriteString("?=\"\r\n\r\n")

			b := make([]byte, base64.StdEncoding.EncodedLen(len(v)))
			base64.StdEncoding.Encode(b, v)

			// write base64 content in lines of up to 76 chars
			for i, l := 0, len(b); i < l; i++ {
				buf.WriteByte(b[i])
				if (i+1)%76 == 0 {
					buf.WriteString("\r\n")
				}
			}

			buf.WriteString("\r\n--" + boundary)
		}
		// for k, v := range m.Attachments {
		// 	buf.WriteString(fmt.Sprintf("\n\n--%s\n", boundary))
		//
		// 	ext := filepath.Ext(k)
		// 	mimetype := mime.TypeByExtension(ext)
		// 	buf.WriteString(fmt.Sprintf("Content-Type: %s\n", mimetype))
		// 	buf.WriteString("Content-Transfer-Encoding: base64\n")
		// 	buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=%s\n", k))
		//
		// 	// b := make([]byte, base64.StdEncoding.EncodedLen(len(v)))
		// 	// base64.StdEncoding.Encode(b, v)
		// 	// buf.Write(b)
		//
		// 	b := make([]byte, base64.StdEncoding.EncodedLen(len(v)))
		// 	base64.StdEncoding.Encode(b, v)
		//
		// 	// write base64 content in lines of up to 76 chars
		// 	for i, l := 0, len(b); i < l; i++ {
		// 		buf.WriteByte(b[i])
		// 		if (i+1)%76 == 0 {
		// 			buf.WriteString("\r\n")
		// 		}
		// 	}
		// 	buf.WriteString(fmt.Sprintf("\n--%s", boundary))
		// }

		buf.WriteString("--")
	}

	return buf.Bytes()
}

func (r *request) AttachFile(src string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	_, fileName := filepath.Split(src)
	r.Attachments[fileName] = b
	return nil
}
