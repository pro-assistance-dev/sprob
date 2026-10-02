package email

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
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

// displayName — человекочитаемое имя отправителя. Голый адрес (тем более
// punycode-домен) в From повышает спам-скор; имя делает письмо легитимнее.
const displayName = "АНО «Просодействие»"

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

// encodeHeader — RFC 2047 для заголовков с не-ASCII (тема письма с кириллицей,
// имена получателей). Без кодирования кириллица уезжает в заголовке как есть и
// ловится спам-фильтрами/превращается в кракозябры.
func encodeHeader(s string) string {
	if s == "" || isASCII(s) {
		return s
	}
	return mime.QEncoding.Encode("utf-8", s)
}

func isASCII(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
	}
	return true
}

// formatAddress — «Имя <addr>» с корректным кодированием имени; голый адрес
// оставляем без изменений.
func formatAddress(name, addr string) string {
	if name == "" {
		return addr
	}
	return fmt.Sprintf("%s <%s>", encodeHeader(name), addr)
}

// messageID — уникальный идентификатор письма. Отсутствие Message-ID —
// заметный спам-признак, а без него часть получателей вообще схлопывает.
func messageID(from string) string {
	domain := "localhost"
	if at := strings.LastIndex(from, "@"); at >= 0 && at < len(from)-1 {
		domain = from[at+1:]
	}
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("<%d@%s>", time.Now().UnixNano(), domain)
	}
	return fmt.Sprintf("<%s@%s>", hex.EncodeToString(b), domain)
}

// buildHeaders — заголовки письма в детерминированном порядке (карта в Go
// итерируется случайно, а порядок заголовков влияет на некоторые антиспамы).
func (e *Email) buildHeaders() string {
	to := make([]string, 0, len(e.request.To))
	for _, t := range e.request.To {
		to = append(to, t)
	}
	pairs := [][2]string{
		{"From", formatAddress(displayName, e.config.From)},
		{"To", strings.Join(to, ", ")},
		{"Subject", encodeHeader(e.request.Subject)},
		{"Date", time.Now().Format(time.RFC1123Z)},
		{"Message-ID", messageID(e.config.From)},
		// List-Unsubscribe — обязателен для массовых рассылок (Gmail/Yandex):
		// без него письмо почти гарантированно уходит в «Промоакции»/«Спам».
		{"List-Unsubscribe", "<mailto:" + e.config.From + "?subject=unsubscribe>"},
		{"List-Unsubscribe-Post", "List-Unsubscribe=One-Click"},
		{"MIME-Version", "1.0"},
		{"Content-Type", "text/html; charset=\"utf-8\""},
		{"Content-Transfer-Encoding", "quoted-printable"},
	}
	var b strings.Builder
	for _, p := range pairs {
		fmt.Fprintf(&b, "%s: %s\r\n", p[0], p[1])
	}
	return b.String()
}

// writeBody — тело письма в quoted-printable.
func (e *Email) writeBody(w *strings.Builder) error {
	enc := quotedprintable.NewWriter(w)
	if _, err := enc.Write([]byte(e.request.Body)); err != nil {
		return err
	}
	return enc.Close()
}

// buildMessage — сырое письмо для SMTP: заголовки + тело (или multipart с
// вложениями).
func (e *Email) buildMessage() (string, error) {
	var msg strings.Builder
	msg.WriteString(e.buildHeaders())
	if len(e.request.Attachments) == 0 {
		msg.WriteString("\r\n")
		if err := e.writeBody(&msg); err != nil {
			return "", err
		}
		return msg.String(), nil
	}

	var body strings.Builder
	if err := e.writeBody(&body); err != nil {
		return "", err
	}

	var buf strings.Builder
	buf.WriteString("From: " + formatAddress(displayName, e.config.From) + "\r\n")
	buf.WriteString("To: " + strings.Join(e.request.To, ", ") + "\r\n")
	buf.WriteString("Subject: " + encodeHeader(e.request.Subject) + "\r\n")
	buf.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
	buf.WriteString("Message-ID: " + messageID(e.config.From) + "\r\n")
	buf.WriteString("MIME-Version: 1.0\r\n")

	writer := multipart.NewWriter(&buf)
	buf.WriteString("Content-Type: multipart/mixed; boundary=" + writer.Boundary() + "\r\n")
	buf.WriteString("\r\n")

	header := make(map[string][]string)
	header["Content-Type"] = []string{"text/html; charset=utf-8"}
	header["Content-Transfer-Encoding"] = []string{"quoted-printable"}
	part, err := writer.CreatePart(header)
	if err != nil {
		return "", err
	}
	if _, err := part.Write([]byte(body.String())); err != nil {
		return "", err
	}

	for name, data := range e.request.Attachments {
		header := make(map[string][]string)
		header["Content-Type"] = []string{mime.TypeByExtension(filepath.Ext(name))}
		header["Content-Transfer-Encoding"] = []string{"base64"}
		header["Content-Disposition"] = []string{"attachment; filename=\"" + name + "\""}
		part, err := writer.CreatePart(header)
		if err != nil {
			return "", err
		}
		if _, err := part.Write(data); err != nil {
			return "", err
		}
	}
	if err := writer.Close(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// sendEmail func
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

	message, err := e.buildMessage()
	if err != nil {
		return err
	}

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
	defer func() { _ = conn.Close() }()

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()

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

	if _, err = w.Write([]byte(message)); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	if err = c.Quit(); err != nil {
		return err
	}

	if e.config.WriteTestFile {
		if err := os.WriteFile("./application-generate_send.html", []byte(message), 0o600); err != nil {
			log.Printf("Error writing test file: %v", err)
		}
	}

	return nil
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
