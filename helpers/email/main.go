package email

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	gomail "gopkg.in/gomail.v2"

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
	m := gomail.NewMessage()
	m.SetHeader("From", e.config.From)
	m.SetHeader("To", e.request.To...)
	m.SetHeader("Subject", e.request.Subject)
	m.SetBody("text/html", e.request.Body)

	d := gomail.NewDialer(
		e.config.Server,
		e.config.Port,
		e.config.From,
		e.config.Password,
	)
	d.SSL = false
	d.TLSConfig = &tls.Config{ServerName: e.config.Server}

	if err := d.DialAndSend(m); err != nil {
		return err
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
