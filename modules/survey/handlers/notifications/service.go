package notifications

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/survey/models"
)

// Service — редактор уведомлений публикации.
type Service struct {
	r *Repository
}

// GetAll — уведомления публикации.
func (s *Service) GetAll(c context.Context, publicationID string) (models.Notifications, error) {
	return s.r.GetAll(c, publicationID)
}

// ReplaceAll — заменить уведомления публикации целиком.
func (s *Service) ReplaceAll(c context.Context, publicationID string, items models.Notifications) (models.Notifications, error) {
	if err := s.r.ReplaceAll(c, publicationID, items); err != nil {
		return nil, err
	}
	return s.r.GetAll(c, publicationID)
}

// GetLogs — журнал отправок публикации (для таба «Уведомления»).
func (s *Service) GetLogs(c context.Context, publicationID string) (models.NotificationLogs, error) {
	return s.r.GetLogs(c, publicationID)
}
