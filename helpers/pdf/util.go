package pdf

import (
	"net/http"
	"os"
	"reflect"
)

func inArray(val interface{}, array interface{}) bool {
	return atArray(val, array) != -1
}

func atArray(val interface{}, array interface{}) (index int) {
	index = -1

	if reflect.TypeOf(array).Kind() == reflect.Slice {
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) {
				index = i
				return
			}
		}
	}

	return
}

func getMimeType(file *os.File) (string, error) {
	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)
	_, readError := file.Read(buffer)
	if readError != nil {
		return "error", readError
	}

	// Always returns a valid content-type and "application/octet-stream" if no others seemed to match.
	return http.DetectContentType(buffer), nil
}
