package metabase

type Card struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type Cards []*Card
