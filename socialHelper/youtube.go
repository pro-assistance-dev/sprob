package socialHelper

import (
	"encoding/json"
	"log"
	"mdgkb/mdgkb-server/models"
	"net/http"
)

type youTubeElement struct {
	YouTubeID youTubeID      `json:"id"`
	Snippet   youTubeSnippet `json:"snippet"`
}
type youTubeElements []*youTubeElement

type youTubeID struct {
	VideoID string `json:"videoId"`
}

type youTubeSnippet struct {
	YouTubeThumbnails youTubeThumbnails `json:"thumbnails"`
	Description       string            `json:"description"`
	Title             string            `json:"title"`
}

type youTubeThumbnails struct {
	Medium youTubeMedium `json:"medium"`
}

type youTubeMedium struct {
	Url string `json:"url"`
}

type youTubeStruct struct {
	Items youTubeElements `json:"items"`
}

func (i *youTubeStruct) getWebFeed(data *http.Response) models.Socials {
	i.decode(data)
	socials := make(models.Socials, 0)
	for index := range i.Items {
		item := models.Social{
			Type:        models.SocialTypeYouTube,
			Description: i.Items[index].Snippet.Description,
			Link:        "https://www.youtube.com/watch?v=" + i.Items[index].YouTubeID.VideoID,
			Image:       i.Items[index].Snippet.YouTubeThumbnails.Medium.Url,
			MediaType:   models.MediaTypeImage,
		}
		socials = append(socials, &item)
	}
	return socials
}

func (i *youTubeStruct) decode(data *http.Response) {
	err := json.NewDecoder(data.Body).Decode(&i)
	if err != nil {
		log.Println(err)
	}
	data.Body.Close()
}
