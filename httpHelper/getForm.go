package httpHelper

import (
	"encoding/json"
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

func (i *HTTPHelper) GetForm(c *gin.Context, item interface{}) (map[string][]*multipart.FileHeader, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(form.Value["form"][0]), &item)
	if err != nil {
		return nil, err
	}
	return form.File, nil
}
