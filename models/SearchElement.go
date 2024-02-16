package models

type SearchElement struct {
	Description        string             `json:"description"`
	Value              string             `json:"value"`
	Label              string             `json:"label"`
	Route              string             `json:"route"`
	SearchGroup        *SearchGroup       `json:"searchGroup"`
	SearchElementMetas SearchElementMetas `json:"searchElementMetas"`
}

type SearchElements []*SearchElement
