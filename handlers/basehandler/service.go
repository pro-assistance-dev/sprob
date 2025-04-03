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

func (s *Service[T]) Get(c context.Context, id string) (T, error) {
	return s.R.Get(c, id)
}

func (s *Service[T]) Update(c context.Context, item *T) error {
	return s.R.Update(c, item)
}

func (s *Service[T]) UpdateMany(c context.Context, items []*T) error {
	return s.R.UpdateMany(c, items)
}

func (s *Service[T]) Delete(c context.Context, id string) error {
	return s.R.Delete(c, id)
}
