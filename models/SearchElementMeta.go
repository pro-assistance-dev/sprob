package models

type SearchElementMeta struct {
	Name                  string                 `json:"name"`
	Value                 string                 `json:"value"`
	SearchGroupMetaColumn *SearchGroupMetaColumn `json:"searchGroupMetaColumn"`
}

type SearchElementMetas []*SearchElementMeta
