package search

import (
	"context"
	"fmt"

	"github.com/pro-assistance-dev/sprob/models"
)

func (s *Service) Search(c context.Context, searchModel *models.SearchModel) (err error) {
	searchGroup, err := s.r.GetGroupByKey(c, searchModel.Key)
	if err != nil {
		return err
	}
	searchModel.SearchGroup = searchGroup
	err = s.r.Search(c, searchModel)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) SearchMain(c context.Context, searchModel *models.SearchModel) (err error) {
	searchModel.SearchGroups, err = s.r.GetGroups(c, searchModel.SearchGroupID)
	if err != nil {
		return err
	}
	fmt.Println("searchModel:", searchModel.SearchGroups)
	for i := range searchModel.SearchGroups {
		err = s.r.Search(c, searchModel)
		if err != nil {
			return err
		}
		searchModel.SearchGroups[i].BuildRoutes()
	}
	return nil
}
