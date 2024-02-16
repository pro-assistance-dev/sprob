package social

import (
	"encoding/json"
	"log"
	"net/http"
)

type youTubeElement struct {
	ID      interface{}    `json:"id"`
	Snippet youTubeSnippet `json:"snippet"`
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

func (i *youTubeStruct) getWebFeed(data *http.Response) Socials {
	i.decode(data)
	socials := make(Socials, 0)
	for index := range i.Items {
		item := Social{
			Type:        SocialTypeYouTube,
			Title:       i.Items[index].Snippet.Title,
			Description: i.Items[index].Snippet.Description,
			Image:       i.Items[index].Snippet.YouTubeThumbnails.Medium.Url,
			MediaType:   MediaTypeImage,
		}
		switch v := i.Items[index].ID.(type) {
		case string:
			item.Link = "https://www.youtube.com/watch?v=" + v
		case map[string]interface{}:
			item.Link = "https://www.youtube.com/watch?v=" + v["videoId"].(string)
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
