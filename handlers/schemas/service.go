package schemas

import (
	"context"

	"github.com/pro-assistance-dev/sprob/helpers/project"
)

func (s *Service) Create(c context.Context, item *project.Schema) error {
	return s.r.Create(c, item)
}

func (s *Service) GetAll(c context.Context) (project.SchemasWithCount, error) {
	return s.r.GetAll(c)
}

func (s *Service) Get(c context.Context, id string) (*project.Schema, error) {
	return s.r.Get(c, id)
}

func (s *Service) Update(c context.Context, item *project.Schema) error {
	return s.r.Update(c, item)
}

func (s *Service) Delete(c context.Context, id string) error {
	return s.r.Delete(c, id)
}
