package elasticSearchHelper

type ElasticSearchHelper struct {
	On bool `json:"on"`
}

func NewElasticSearchHelper(on bool) *ElasticSearchHelper {
	return &ElasticSearchHelper{On: on}
}
