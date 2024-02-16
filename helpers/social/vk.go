package social

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
)

type vkItem struct {
	ID          int           `json:"id"`
	FromID      int           `json:"from_id"`
	Text        string        `json:"text"`
	Attachments vkAttachments `json:"attachments"`
}
type vkItems []*vkItem

func (i *vkItem) getPhotoURL() string {
	if len(i.Attachments) == 0 {
		return ""
	}
	attachment := i.Attachments[0]
	url := ""
	for _, size := range attachment.Photo.Sizes {
		if size.Type == "q" {
			url = size.Url
			break
		}
	}
	return url
}

type vkAttachment struct {
	Photo vkPhoto `json:"photo"`
}

type vkAttachments []*vkAttachment

type vkPhoto struct {
	Sizes vkSizes `json:"sizes"`
}

type vkSize struct {
	Type string `json:"type"`
	Url  string `json:"url"`
}

type vkSizes []*vkSize

type vkStruct struct {
	Response vkResponse `json:"response"`
}

type vkResponse struct {
	Items vkItems `json:"items"`
}

func (i *vkStruct) getWebFeed(data *http.Response) Socials {
	i.decode(data)
	socials := make(Socials, 0)
	for _, item := range i.Response.Items {
		title := item.Text
		if len(item.Text) > 100 {
			title = item.Text[:100]
		}
		emojiRx := regexp.MustCompile(`[\x{1F600}-\x{1F6FF}|[\x{2600}-\x{26FF}]`)
		title = emojiRx.ReplaceAllString(title, "")

		item := Social{
			Type:        SocialTypeVK,
			Title:       title,
			Link:        fmt.Sprintf("https://vk.com/morozdgkbdzm?w=wall%d_%d", item.FromID, item.ID),
			Description: item.Text,
			Image:       item.getPhotoURL(),
			MediaType:   MediaTypeImage,
		}
		if len(item.Title) > 0 && len(item.Image) > 0 {
			socials = append(socials, &item)
		}
	}
	return socials
}

func (i *vkStruct) decode(data *http.Response) {
	err := json.NewDecoder(data.Body).Decode(&i)
	if err != nil {
		log.Println(err)
	}
	data.Body.Close()
}
