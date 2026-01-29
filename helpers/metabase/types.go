package metabase

// QueryResult результат выполнения запроса
type QueryResult struct {
	Data struct {
		Columns []string        `json:"columns"`
		Rows    [][]interface{} `json:"rows"`
	} `json:"data"`
}

// Database информация о базе данных
type Database struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Engine      string `json:"engine"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// Question сохраненный вопрос
type Question struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	DatabaseID  int                    `json:"database_id"`
	Collection  map[string]interface{} `json:"collection"`
	Query       map[string]interface{} `json:"query"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

// Collection коллекция вопросов
type Collection struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Slug        string `json:"slug"`
	Color       string `json:"color"`
}

// User пользователь Metabase
type User struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	IsActive  bool   `json:"is_active"`
}

// ErrorResponse ответ с ошибкой
type ErrorResponse struct {
	Message string                 `json:"message"`
	Errors  map[string]interface{} `json:"errors,omitempty"`
}
