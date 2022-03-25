package httpHelper

import "fmt"

func (i *HTTPHelper) GetRestorePasswordURL(userId string, uniqueId string) string {
	return fmt.Sprintf("%s/restore/password/%s/%s", i.Host, userId, uniqueId)
}
