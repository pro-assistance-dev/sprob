package httpHelper

import "fmt"

func (i *HTTPHelper) GetRestorePasswordURL(userId string, uniqueId string, domen string) (string, error) {
	return fmt.Sprintf("%s/restore/password/%s/%s", domen, userId, uniqueId), nil
}
