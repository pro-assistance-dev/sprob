package helper

type Mode string

const (
	Migrate Mode = "migrate"
	Run          = "run"
	Dump         = "dump"
	Test         = "test"
)
