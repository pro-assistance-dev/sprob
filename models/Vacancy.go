package models

// import (
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/uptrace/bun"
// )

// type Vacancy struct {
// 	bun.BaseModel     `bun:"vacancies,select:vacancies_view,alias:vacancies_view"`
// 	ID                uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
// 	Title             string    `json:"title"`
// 	Slug              string    `json:"slug"`
// 	Specialization    string    `json:"specialization"`
// 	MinSalary         int       `json:"minSalary"`
// 	MaxSalary         int       `json:"maxSalary"`
// 	SalaryComment     string    `json:"salaryComment"`
// 	Active            bool      `json:"active"`
// 	Experience        string    `json:"experience"`
// 	Schedule          string    `json:"schedule"`
// 	Date              time.Time `bun:"vacancy_date" json:"date"`
// 	ResponsesCount    int       `json:"responsesCount"`
// 	NewResponsesCount int       `json:"newResponsesCount"`
// }

// type Vacancies []*Vacancy
