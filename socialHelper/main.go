package socialHelper

import (
	"context"
	"fmt"
	"log"
	"mdgkb/mdgkb-server/helpers/config"
	"mdgkb/mdgkb-server/models"
	"net/http"
)

type Social struct {
	config.Social
}

func (i *Social) buildInstagramURL() string {
	instagramApi := "https://graph.instagram.com"
	fields := "id,media_url,media_type,thumbnail_url,permalink,caption"
	return fmt.Sprintf("%s/%s/media?fields=%s&access_token=%s", instagramApi, i.InstagramID, fields, i.InstagramToken)
}

func (i *Social) buildYouTubeURL() string {
	const youTubeApi = "https://www.googleapis.com/youtube/v3/search"
	options := "&part=snippet&maxResults=6&order=date&type=video"
	return fmt.Sprintf("%s?key=%s&channelId=%s%s", youTubeApi, i.YouTubeApiKey, i.YouTubeChannelID, options)
}

type SocialType string

const (
	Instagram SocialType = "Instagram"
)

func NewSocial(config config.Social) *Social {
	return &Social{config}
}

func (i *Social) sendRequest(url string) *http.Response {
	ctx := context.Background()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Println(err)
	}
	c := &http.Client{}
	resp, err := c.Do(request)
	if err != nil {
		log.Println(err)
	}
	return resp
}

func (i *Social) GetWebFeed() models.Socials {
	//instagram := instagramStruct{}
	//resp := i.sendRequest(i.buildInstagramURL())
	//instagram.decode(resp)
	youTube := youTubeStruct{}
	socials := youTube.getWebFeed(i.sendRequest(i.buildYouTubeURL()))
	return socials
}
