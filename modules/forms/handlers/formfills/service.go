package formfills

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/forms/models"
)

func (s *Service) Create(c context.Context, item *models.FormFill) error {
	err := R.Create(c, item)
	if err != nil {
		return err
	}
	// item.SetIDForChildren()
	// err = FormFillgroup.CreateService(s.helper).UpsertMany(item.FormFillGroups)
	//err = FormFilldiagnosis.CreateService(s.helper).CreateMany(item.FormFillDiagnosis)
	//if err != nil {
	//	return err
	//}
	return err
}

func (s *Service) GetAll(c context.Context) (models.FormFillsWithCount, error) {
	return R.GetAll(c)
}

func (s *Service) Get(c context.Context, id string) (*models.FormFill, error) {
	item, err := R.Get(c, id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) Update(c context.Context, item *models.FormFill) error {
	err := R.Update(c, item)
	if err != nil {
		return err
	}
	// item.SetIDForChildren()

	//FormFillGroupService := FormFillgroup.CreateService(s.helper)
	//err = FormFillGroupService.UpsertMany(item.FormFillGroups)
	//if err != nil {
	//	return err
	//}
	//err = FormFillGroupService.DeleteMany(item.formfillsForDelete)
	//if err != nil {
	//	return err
	//}

	//FormFillDiagnosisService := FormFilldiagnosis.CreateService(s.helper)
	//err = FormFillDiagnosisService.UpsertMany(item.FormFillDiagnosis)
	//if err != nil {
	//	return err
	//}
	//err = FormFillDiagnosisService.DeleteMany(item.FormFillDiagnosisForDelete)
	//if err != nil {
	//	return err
	//}
	return err
}

func (s *Service) Delete(c context.Context, id *string) error {
	return R.Delete(c, id)
}

// func (s *Service) GetFormFillAndPatient(c context.Context, researchID string, patientID string) (*models.FormFill, *models.Patient, error) {
// 	research, err := R.Get(c, researchID)
// 	if err != nil {
// 		return nil, nil, err
// 	}
//
// 	// patient, err := patients.S.Get(c, patientID)
// 	// if err != nil {
// 	// 	return nil, nil, err
// 	// }
// 	return research, nil, nil
// }
