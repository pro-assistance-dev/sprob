package basehandler

import (
	"context"
)

func (s *Service[T]) Create(c context.Context, item *T) error {
	return s.R.Create(c, item)
}

func (s *Service[T]) GetAll(c context.Context) (any, error) {
	return s.R.GetAll(c)
}

func (s *Service[T]) LabelValue(c context.Context, labelCol string, valueCol string) ([]*LabelValue, error) {
	if labelCol == "" {
		labelCol = "name"
	}
	if valueCol == "" {
		valueCol = "id"
	}
	return s.R.LabelValue(c, labelCol, valueCol)
}

func (s *Service[T]) Get(c context.Context, id string) (T, error) {
	return s.R.Get(c, id)
}

func (s *Service[T]) GetBySlug(c context.Context, slug string) (T, error) {
	return s.R.Get(c, slug)
}

func (s *Service[T]) Update(c context.Context, item *T) error {
	return s.R.Update(c, item)
}

func (s *Service[T]) UpdateMany(c context.Context, items []*T) error {
	return nil
}

func (s *Service[T]) Delete(c context.Context, id string) error {
	return s.R.Delete(c, id)
}
