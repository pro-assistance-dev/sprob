package f

type Operator string

const (
	Eq      Operator = "="
	Ne      Operator = "!="
	Gt      Operator = ">"
	Ge      Operator = "<"
	Btw     Operator = "between"
	Like    Operator = "like"
	In      Operator = "in"
	Null    Operator = "is null"
	NotNull Operator = "is not null"
)
