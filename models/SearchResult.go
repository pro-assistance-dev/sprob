package models

type SearchResult struct {
	Description string `json:"description"`
	Value       string `json:"value"`
	Label       string `json:"label"`
	Route       string `json:"route"`
	Key         string `json:"key"`
	// SearchGroup *SearchGroup `json:"searchGroup"`
	// SearchResultMetas SearchResultMetas `json:"searchResultMetas"`
}

type SearchResults []*SearchResult
