package socialHelper

import (
	"encoding/json"
	"log"
	"net/http"
)

type instagramElement struct {
	Type          SocialType `json:"type"`
	Caption       string     `json:"caption"`
	Permalink     string     `json:"permalink"`
	MediaUrlSnake string     `json:"media_url"`
	MediaUrlCamel string     `json:"mediaUrl"`
	ThumbnailUrl  string     `json:"thumbnail_url"`
	MediaType     MediaType  `json:"media_type"`
}

func (i *instagramElement) setMediaSRC() {
	if i.MediaType == MediaTypeImage {
		i.MediaUrlCamel = i.MediaUrlSnake
	}
	if i.MediaType == MediaTypeVideo {
		i.MediaUrlCamel = i.ThumbnailUrl
	}
}

type instagramElements []*instagramElement

type instagramStruct struct {
	InstagramElements instagramElements `json:"data"`
}

func (i *instagramStruct) decode(data *http.Response) {
	err := json.NewDecoder(data.Body).Decode(&data)
	if err != nil {
		log.Println(err)
	}
	data.Body.Close()

	for index := range i.InstagramElements {
		i.InstagramElements[index].setMediaSRC()
	}
}
