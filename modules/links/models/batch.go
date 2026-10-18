package models

// BatchItem — один элемент пакетного выпуска («выпустить для этажа/отделения»).
type BatchItem struct {
	Title      string `json:"title"`
	TargetType string `json:"targetType"`
	TargetID   string `json:"targetId"`
	Path       string `json:"path"`
}

// BatchResult — отчёт по пакету (создано / уже было / ошибки).
type BatchResult struct {
	Created int               `json:"created"`
	Existed int               `json:"existed"`
	Failed  []BatchItemResult `json:"failed"`
}

type BatchItemResult struct {
	TargetID string `json:"targetId"`
	Error    string `json:"error"`
}
