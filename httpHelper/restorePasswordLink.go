package httpHelper

import "fmt"

func (i *HTTPHelper) GetRestorePasswordURL(userId string, uniqueId string) string {
	return fmt.Sprintf("%s:%s/restore/password/%s/%s", i.Host, i.Port, userId, uniqueId)
}
