package utilHelper

type UtilHelper struct {
	BinPath string
}

func NewUtilHelper(binPath string) *UtilHelper {
	return &UtilHelper{BinPath: binPath}
}
