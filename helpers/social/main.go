package social

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/pro-assistance-dev/sprob/config"
)

type SocialModel struct { //nolint:golint
	Type        SocialType `json:"type"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Link        string     `json:"link"`
	Image       string     `json:"image"`
	MediaType   MediaType  `json:"mediaType"`
}

type Socials []*SocialModel //nolint:golint

type SocialData struct { //nolint:golint
	Socials Socials `json:"data"`
}

type SocialType string //nolint:golint

const (
	SocialTypeInstagram SocialType = "Instagram"
	SocialTypeYouTube   SocialType = "YouTube"
	SocialTypeVK        SocialType = "VK"
)

type MediaType string

const (
	MediaTypeImage         MediaType = "IMAGE"
	MediaTypeVideo         MediaType = "VIDEO"
	MediaTypeCarouselAlbum MediaType = "CAROUSEL_ALBUM"
)

type Social struct {
	config.Social
}

// func (i *Social) buildInstagramURL() string {
// 	instagramAPI := "https://graph.instagram.com"
// 	fields := "id,media_url,media_type,thumbnail_url,permalink,caption"
// 	return fmt.Sprintf("%s/%s/media?fields=%s&access_token=%s", instagramAPI, i.InstagramID, fields, i.InstagramToken)
// }

const (
	youTubeAPIV3 = "https://www.googleapis.com/youtube/v3/"
	vkAPI        = "https://api.vk.com/method/wall.get"
)

func (i *Social) buildVkURL() string {
	const vkWallAPI = "https://api.vk.com/method/wall.get"
	queryParams := "?filter=owner&count=10&v=5.131&owner_id=%s&access_token=%s"
	return fmt.Sprintf(vkWallAPI+queryParams, i.VkGroupID, i.VkServiceApplicationKey)
}

func (i *Social) buildYouTubeChannelURL() string {
	const youTubeAPI = "https://www.googleapis.com/youtube/v3/search"
	options := "&part=snippet&maxResults=6&order=date&type=video"
	return fmt.Sprintf("%s?key=%s&channelId=%s%s", youTubeAPI, i.YouTubeAPIKey, i.YouTubeChannelID, options)
}

func (i *Social) buildYouTubeVideosURL(idPool []string) string {
	options := "videos?part=id%2C+snippet"
	urlSource, err := url.Parse(youTubeAPIV3 + options)
	if err != nil {
		return ""
	}
	q := urlSource.Query()
	for _, id := range idPool {
		q.Add("id", id)
	}
	urlSource.RawQuery = q.Encode()
	return fmt.Sprintf("%s&key=%s", urlSource.String(), i.YouTubeAPIKey)
}

func NewSocial(social config.Social) *Social {
	return &Social{social}
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

func (i *Social) GetWebFeed() Socials {
	var socials Socials

	// instagram := instagramStruct{}
	// resp := i.sendRequest(i.buildInstagramURL())
	// instagram.decode(resp)

	youTube := youTubeStruct{}
	socialsYouTube := youTube.getWebFeed(i.sendRequest(i.buildYouTubeChannelURL()))
	socials = append(socials, socialsYouTube...)

	vk := vkStruct{}
	socialsVk := vk.getWebFeed(i.sendRequest(i.buildVkURL()))
	socials = append(socials, socialsVk...)

	return socials
}

func (i *Social) GetYouTubeVideosInfo(idPool []string) Socials {
	youTube := youTubeStruct{}
	socials := youTube.getWebFeed(i.sendRequest(i.buildYouTubeVideosURL(idPool)))
	return socials
}
