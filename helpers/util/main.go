package util

type Util struct {
	BinPath string
}

func NewUtil(binPath string) *Util {
	return &Util{BinPath: binPath}
}
