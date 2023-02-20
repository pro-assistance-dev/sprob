package socialHelper

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/pro-assistance/pro-assister/config"
)

type Social struct {
	Type        SocialType `json:"type"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Link        string     `json:"link"`
	Image       string     `json:"image"`
	MediaType   MediaType  `json:"mediaType"`
}

type Socials []*Social

type SocialData struct {
	Socials Socials `json:"data"`
}

type SocialType string

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

type SocialHelper struct {
	config.Social
}

func (i *SocialHelper) buildInstagramURL() string {
	instagramApi := "https://graph.instagram.com"
	fields := "id,media_url,media_type,thumbnail_url,permalink,caption"
	return fmt.Sprintf("%s/%s/media?fields=%s&access_token=%s", instagramApi, i.InstagramID, fields, i.InstagramToken)
}

const youTubeApiV3 = "https://www.googleapis.com/youtube/v3/"
const vkApi = "https://api.vk.com/method/wall.get"

func (i *SocialHelper) buildVkURL() string {
	const vkWallApi = "https://api.vk.com/method/wall.get"
	queryParams := "?filter=owner&count=10&v=5.131&owner_id=%s&access_token=%s"
	fmt.Println(fmt.Sprintf(vkWallApi+queryParams, i.VkGroupID, i.VkServiceApplicationKey))
	return fmt.Sprintf(vkWallApi+queryParams, i.VkGroupID, i.VkServiceApplicationKey)
}

func (i *SocialHelper) buildYouTubeChannelURL() string {
	const youTubeApi = "https://www.googleapis.com/youtube/v3/search"
	options := "&part=snippet&maxResults=6&order=date&type=video"
	return fmt.Sprintf("%s?key=%s&channelId=%s%s", youTubeApi, i.YouTubeApiKey, i.YouTubeChannelID, options)
}

func (i *SocialHelper) buildYouTubeVideosURL(idPool []string) string {
	options := "videos?part=id%2C+snippet"
	urlSource, err := url.Parse(youTubeApiV3 + options)
	if err != nil {
		return ""
	}
	q := urlSource.Query()
	for _, id := range idPool {
		q.Add("id", id)
	}
	urlSource.RawQuery = q.Encode()
	return fmt.Sprintf("%s&key=%s", urlSource.String(), i.YouTubeApiKey)
}

func NewSocial(social config.Social) *SocialHelper {
	return &SocialHelper{social}
}

func (i *SocialHelper) sendRequest(url string) *http.Response {
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

func (i *SocialHelper) GetWebFeed() Socials {
	var socials Socials

	//instagram := instagramStruct{}
	//resp := i.sendRequest(i.buildInstagramURL())
	//instagram.decode(resp)

	youTube := youTubeStruct{}
	socialsYouTube := youTube.getWebFeed(i.sendRequest(i.buildYouTubeChannelURL()))
	socials = append(socials, socialsYouTube...)

	vk := vkStruct{}
	socialsVk := vk.getWebFeed(i.sendRequest(i.buildVkURL()))
	socials = append(socials, socialsVk...)

	return socials
}

func (i *SocialHelper) GetYouTubeVideosInfo(idPool []string) Socials {
	youTube := youTubeStruct{}
	socials := youTube.getWebFeed(i.sendRequest(i.buildYouTubeVideosURL(idPool)))
	return socials
}
