package http

import "fmt"

func (i *HTTP) GetRestorePasswordURL(userID string, uniqueID string) string {
	return fmt.Sprintf("%s/restore/password/%s/%s", i.Host, userID, uniqueID)
}
