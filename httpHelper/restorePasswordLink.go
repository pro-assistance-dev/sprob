package httpHelper

import "fmt"

func (i *HTTPHelper) GetRestorePasswordURL(userId string, uniqueId string) (string, error) {
	ip, err := i.GetMyIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://%s/restore/password/%s/%s", ip.String(), userId, uniqueId), nil
}
