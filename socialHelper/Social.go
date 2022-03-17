package socialHelper

type Social struct {
	Type        SocialType `json:"type"`
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
	Instagram SocialType = "Instagram"
)

const (
	SocialTypeInstagram SocialType = "Instagram"
	SocialTypeYouTube   SocialType = "YouTube"
)

type MediaType string

const (
	MediaTypeImage         MediaType = "IMAGE"
	MediaTypeVideo         MediaType = "VIDEO"
	MediaTypeCarouselAlbum MediaType = "CAROUSEL_ALBUM"
)
