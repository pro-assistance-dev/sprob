package helper

type Mode string

const (
	Migrate Mode = "migrate"
	Run     Mode = "run"
	Dump    Mode = "dump"
	Listen  Mode = "listen"
	Test    Mode = "test"
)
