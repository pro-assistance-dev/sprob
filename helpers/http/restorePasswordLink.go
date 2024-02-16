package http

import "fmt"

func (i *HTTP) GetRestorePasswordURL(userId string, uniqueId string) string {
	return fmt.Sprintf("%s/restore/password/%s/%s", i.Host, userId, uniqueId)
}
