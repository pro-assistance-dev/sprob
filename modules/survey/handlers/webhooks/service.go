package webhooks

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/survey/models"
)

// Service — редактор вебхуков публикации.
type Service struct {
	r *Repository
}

// GetAll — вебхуки публикации.
func (s *Service) GetAll(c context.Context, publicationID string) (models.Webhooks, error) {
	return s.r.GetAll(c, publicationID)
}

// ReplaceAll — заменить вебхуки публикации целиком.
func (s *Service) ReplaceAll(c context.Context, publicationID string, items models.Webhooks) (models.Webhooks, error) {
	if err := s.r.ReplaceAll(c, publicationID, items); err != nil {
		return nil, err
	}
	return s.r.GetAll(c, publicationID)
}
