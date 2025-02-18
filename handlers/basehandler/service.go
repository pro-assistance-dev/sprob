package basehandler

import (
	"context"
)

func (s *Service[TSingle, TPlural, TPluralWithCount]) Create(c context.Context, item *TSingle) error {
	return s.R.Create(c, item)
}

// func (s *Service[TSingle, TPlural, TPluralWithCount]) GetAll(c context.Context) (TPluralWithCount, error) {
// 	return s.R.GetAll(c)
// }
//
// func (s *Service[TSingle, TPlural, TPluralWithCount]) Get(c context.Context, id string) (*models.Event, error) {
// 	return s.R.Get(c, id)
// }
//
// func (s *Service[TSingle, TPlural, TPluralWithCount]) Update(c context.Context, item *models.Event) error {
// 	return s.R.Update(c, item)
// }
//
// func (s *Service[TSingle, TPlural, TPluralWithCount]) Delete(c context.Context, id string) error {
// 	return s.R.Delete(c, id)
// }
