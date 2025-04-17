package email

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"mime"
	"mime/multipart"
	"net"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"

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
			e.request.AttachFile(f)
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
	body := e.request.ToBytes(e.config.From)
	// header := map[string]string{}
	// header["To"] = strings.Join(e.request.To, ",")
	// header["From"] = e.config.From
	// header["Subject"] = e.request.Subject
	// header["MIME-Version"] = "1.0"
	// header["Content-Type"] = "text/html; charset=\"utf-8\""
	// // header["Content-Transfer-Encoding"] = "base64"
	// message := ""
	// for k, v := range header {
	// 	message += fmt.Sprintf("%s: %s\r\n", k, v)
	// }
	// // message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(e.request.Body))
	// message += "\r\n" + e.request.Body
	//
	// boundary := "\r"
	// if e.request.Attachments.len > 0 {
	// 	for k, v := range e.request.Attachments {
	// 		buf.WriteString(fmt.Sprintf("\n\n--%s\n", boundary))
	// 		buf.WriteString(fmt.Sprintf("Content-Type: %s\n", http.DetectContentType(v)))
	// 		buf.WriteString("Content-Transfer-Encoding: base64\n")
	// 		buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=%s\n", k))
	//
	// 		b := make([]byte, base64.StdEncoding.EncodedLen(len(v)))
	// 		base64.StdEncoding.Encode(b, v)
	// 		buf.Write(b)
	// 		buf.WriteString(fmt.Sprintf("\n--%s", boundary))
	// 	}
	//
	// 	buf.WriteString("--")
	// }

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
	_, err = w.Write(body)
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
		err = os.WriteFile("./application-generate_send.html", []byte(body), 0600)
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func (m *request) ToBytes(from string) []byte {
	buf := bytes.NewBuffer(nil)
	withAttachments := len(m.Attachments) > 0
	buf.WriteString(fmt.Sprintf("Subject: %s\n", m.Subject))
	buf.WriteString(fmt.Sprintf("To: %s\n", strings.Join(m.To, ",")))
	buf.WriteString(fmt.Sprintf("From: %s\n", strings.Join([]string{from}, ",")))
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
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\n", boundary))
		buf.WriteString(fmt.Sprintf("--%s\n", boundary))
	} else {
		buf.WriteString("Content-Type: text/plain; charset=utf-8\n")
	}

	buf.WriteString(fmt.Sprintf("\n\n--%s\n", boundary))
	buf.WriteString(fmt.Sprintf("Content-Type: text/html; charset=\"utf-8\" boundary=%s\n", boundary))
	buf.WriteString(m.Body)
	buf.WriteString(fmt.Sprintf("\n--%s", boundary))

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

// Bytes returns the mail data
func (m *request) Bytes() []byte {
	buf := bytes.NewBuffer(nil)

	buf.WriteString("From: " + m.From + "\r\n")

	t := time.Now()
	buf.WriteString("Date: " + t.Format(time.RFC1123Z) + "\r\n")

	buf.WriteString("To: " + strings.Join(m.To, ",") + "\r\n")
	// if len(m.Cc) > 0 {
	// 	buf.WriteString("Cc: " + strings.Join(m.Cc, ",") + "\r\n")
	// }

	// fix  Encode
	coder := base64.StdEncoding
	subject := "=?UTF-8?B?" + coder.EncodeToString([]byte(m.Subject)) + "?="
	buf.WriteString("Subject: " + subject + "\r\n")

	// if len(m.ReplyTo) > 0 {
	// 	buf.WriteString("Reply-To: " + m.ReplyTo + "\r\n")
	// }

	buf.WriteString("MIME-Version: 1.0\r\n")

	// Add custom headers
	if len(m.Headers) > 0 {
		for _, header := range m.Headers {
			buf.WriteString(fmt.Sprintf("%s: %s\r\n", header.Key, header.Value))
		}
	}

	boundary := "f46d043c813270fc6b04c2d223da"

	if len(m.Attachments) > 0 {
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
