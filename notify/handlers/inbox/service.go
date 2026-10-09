package inbox

import (
	"context"

	"github.com/pro-assistance-dev/sprob/notify/models"
)

func (s *Service) GetAll(c context.Context, userID, group string) (models.InboxWithCount, error) {
	return s.r.GetAll(c, userID, group)
}

func (s *Service) MarkRead(c context.Context, id string) error {
	return s.r.MarkRead(c, id)
}

func (s *Service) MarkAllRead(c context.Context, userID, group string) error {
	return s.r.MarkAllRead(c, userID, group)
}
