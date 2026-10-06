package invites

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pro-assistance-dev/sprob/modules/survey/models"
	"github.com/pro-assistance-dev/sprob/modules/survey/notify"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Repository — создание/чтение приглашений публикации.
type Repository struct {
	db *bun.DB
}

// GetPublication — публикация со связями (нужны slug/title/params для письма).
func (r *Repository) GetPublication(c context.Context, id string) (*models.Publication, error) {
	item := models.Publication{}
	err := r.db.NewSelect().
		Model(&item).
		Relation("Params", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("survey_params.sort_order")
		}).
		Where("publications.id = ?", id).
		Scan(c)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetByEmails — уже существующие приглашения публикации по email (чтобы повторная
// рассылка не плодила дубликаты).
func (r *Repository) GetByEmails(c context.Context, publicationID uuid.NullUUID, emails []string) (models.Invites, error) {
	items := make(models.Invites, 0)
	if len(emails) == 0 {
		return items, nil
	}
	err := r.db.NewSelect().
		Model(&items).
		Where("invites.publication_id = ?", publicationID).
		Where("lower(invites.email) in (?)", bun.In(lowerAll(emails))).
		Scan(c)
	return items, err
}

// Insert — сохраняет приглашения пачкой.
func (r *Repository) Insert(c context.Context, items models.Invites) error {
	if len(items) == 0 {
		return nil
	}
	_, err := r.db.NewInsert().Model(&items).Exec(c)
	return err
}

// Service — рассылка приглашений: создаёт персональные ссылки и отправляет email.
type Service struct {
	r      *Repository
	sender func(to []string, subject, body string) error
}

// NewService — с штатным SMTP-отправителем (EMAIL_* из окружения).
func NewService(db *bun.DB) *Service {
	return &Service{r: &Repository{db: db}, sender: notify.SendEmail}
}

// SendResult — итог по одному адресату.
type SendResult struct {
	Email string `json:"email"`
	Fio   string `json:"fio"`
	Link  string `json:"link"`
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// Recipient — ввод: кому и (необязательно) как обращаться.
type Recipient struct {
	Email string `json:"email"`
	Fio   string `json:"fio"`
}

// Send создаёт приглашения и отправляет письма со ссылками.
// Повторный адрес не дублируется — переиспользуется существующий токен
// (идемпотентно: письмо можно переслать). Невалидные адреса пропускаются с ошибкой.
func (s *Service) Send(c context.Context, publicationID string, recipients []Recipient, subject, body string) ([]SendResult, error) {
	publication, err := s.r.GetPublication(c, publicationID)
	if err != nil {
		return nil, err
	}

	valid := make([]Recipient, 0, len(recipients))
	emails := make([]string, 0, len(recipients))
	for _, r := range recipients {
		addr := strings.TrimSpace(r.Email)
		if !looksLikeEmail(addr) {
			continue
		}
		valid = append(valid, Recipient{Email: addr, Fio: strings.TrimSpace(r.Fio)})
		emails = append(emails, addr)
	}

	existing, err := s.r.GetByEmails(c, publication.ID, emails)
	if err != nil {
		return nil, err
	}
	byEmail := make(map[string]*models.Invite, len(existing))
	for _, inv := range existing {
		byEmail[strings.ToLower(inv.Email)] = inv
	}

	toInsert := make(models.Invites, 0, len(valid))
	pubID := publication.ID
	for _, r := range valid {
		if _, ok := byEmail[strings.ToLower(r.Email)]; ok {
			continue
		}
		inv := &models.Invite{
			ID:            uuid.NullUUID{UUID: uuid.New(), Valid: true},
			PublicationID: pubID,
			Token:         newToken(),
			Email:         r.Email,
			Fio:           r.Fio,
		}
		toInsert = append(toInsert, inv)
		byEmail[strings.ToLower(r.Email)] = inv
	}
	if err := s.r.Insert(c, toInsert); err != nil {
		return nil, err
	}

	base := publicBaseURL()
	results := make([]SendResult, 0, len(valid))
	// ⚠️ Рассылку ведём в фоне: длинный список (сотни адресов) — это минуты SMTP,
	// а HTTP-запрос/канал может оборваться. Сервер продолжит отправку сам;
	// клиент сразу получает ссылки. Статус отправки — в логах сервиса.
	pending := make([]pendingMail, 0, len(valid))
	for _, r := range valid {
		inv := byEmail[strings.ToLower(r.Email)]
		link := fmt.Sprintf("%s/f/%s?invite=%s", base, publication.Slug, inv.Token)
		pending = append(pending, pendingMail{email: r.Email, fio: r.Fio, inv: inv, link: link})
		results = append(results, SendResult{Email: r.Email, Fio: r.Fio, Link: link, OK: true})
	}
	go s.sendAll(pending, publication, subject, body)
	return results, nil
}

// pendingMail — подготовленное письмо (адрес + приглашение + ссылка).
type pendingMail struct {
	email string
	fio   string
	inv   *models.Invite
	link  string
}

// sendAll — фоновая рассылка (не рвётся при обрыве канала клиента).
func (s *Service) sendAll(items []pendingMail, publication *models.Publication, subject, body string) {
	for _, m := range items {
		if err := s.send(m.inv, publication, m.link, subject, body); err != nil {
			log.Printf("[invites] send to %s failed: %v", m.email, err)
		}
	}
}

// send — письмо одному адресату (шаблон с подстановкой ФИО/названия/ссылки).
func (s *Service) send(inv *models.Invite, publication *models.Publication, link, subject, body string) error {
	values := map[string]string{
		"fio":         inv.Fio,
		"publication": publication.Title,
		"link":        link,
	}
	if strings.TrimSpace(subject) == "" {
		subject = "Приглашение пройти опрос: " + publication.Title
	}
	if strings.TrimSpace(body) == "" {
		body = "Здравствуйте{{fio}}!\n\nПриглашаем пройти опрос «{{publication}}».\nСсылка: {{link}}"
	}
	return s.sender([]string{inv.Email}, render(subject, values), render(body, values))
}

// newToken — криптослучайный токен приглашения (32 hex-символа).
func newToken() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

// publicBaseURL — базовый публичный URL для ссылок (SURVEY_PUBLIC_URL из .env).
func publicBaseURL() string {
	return strings.TrimRight(strings.TrimSpace(os.Getenv("SURVEY_PUBLIC_URL")), "/")
}
