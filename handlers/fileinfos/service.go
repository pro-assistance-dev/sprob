package fileinfos

import (
	"context"
	"fmt"
	"pro-assister/models"

	"github.com/google/uuid"
)

func (s *Service) Create(c context.Context, item *models.FileInfo) error {
	if item == nil {
		return nil
	}
	return R.Create(c, item)
}

func (s *Service) Get(c context.Context, id string) (*models.FileInfo, error) {
	return R.Get(c, id)
}

func (s *Service) Update(c context.Context, item *models.FileInfo) error {
	if item == nil {
		return nil
	}
	return R.Update(c, item)
}

func (s *Service) Upsert(c context.Context, item *models.FileInfo) error {
	if item == nil {
		return nil
	}
	fmt.Println(item, item)
	return R.Upsert(c, item)
}

func (s *Service) CreateMany(c context.Context, items models.FileInfos) error {
	if len(items) == 0 {
		return nil
	}
	return R.CreateMany(c, items)
}

func (s *Service) UpsertMany(c context.Context, items models.FileInfos) error {
	if len(items) == 0 {
		return nil
	}
	return R.UpsertMany(c, items)
}

func (s *Service) Delete(c context.Context, id uuid.NullUUID) error {
	return R.Delete(c, id)
}
