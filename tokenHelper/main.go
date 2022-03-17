package tokenHelper

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"os"
	"strings"
	"time"
)

type TokenHelper struct {
	TokenSecret string
}

func NewTokenHelper(tokenSecret string) *TokenHelper {
	return &TokenHelper{TokenSecret: tokenSecret}
}

type AccessDetails struct {
	AccessUuid string
	UserID     string
}

type TokenDetails struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

func (h *TokenHelper) CreateToken(userID string) (*TokenDetails, error) {
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Second).Unix()
	td.AccessUuid = uuid.NewString()

	td.RtExpires = time.Now().Add(time.Minute).Unix()
	td.RefreshUuid = td.AccessUuid + "++" + userID

	var err error
	//Creating Access Token
	os.Setenv("ACCESS_SECRET", h.TokenSecret) //this should be in an env file
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["user_id"] = userID
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(h.TokenSecret))
	if err != nil {
		return nil, err
	}
	//Creating Refresh Token
	os.Setenv("REFRESH_SECRET", h.TokenSecret)
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = userID
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(h.TokenSecret))
	if err != nil {
		return nil, err
	}
	return td, nil
}
func (h *TokenHelper) RefreshToken(refreshToken string) (*TokenDetails, error) {
	token, err := h.VerifyToken(refreshToken)
	fmt.Println(err)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	var userID string
	if ok && token.Valid {
		userID, ok = claims["user_id"].(string)
		if !ok {
			return nil, err
		}
	}
	return h.CreateToken(userID)
}

func (h *TokenHelper) GetUserID(c *gin.Context) (*uuid.UUID, error) {
	accessDetail, err := h.extractTokenMetadata(c.Request)
	if err != nil {
		return nil, err
	}
	uuidFromString, err := uuid.Parse(accessDetail.UserID)
	return &uuidFromString, err
}

func (h *TokenHelper) VerifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.TokenSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (h *TokenHelper) extractTokenMetadata(r *http.Request) (*AccessDetails, error) {
	token, err := h.VerifyToken(h.extractToken(r))
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, err
		}
		userID, ok := claims["user_id"].(string)
		if !ok {
			return nil, err
		}
		return &AccessDetails{
			AccessUuid: accessUuid,
			UserID:     userID,
		}, nil
	}
	return nil, err
}

func (h *TokenHelper) extractToken(r *http.Request) string {
	bearToken := r.Header.Get("token")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return bearToken
}

//
//func (h *TokenHelper)createAuth(userid string, td *TokenDetails, client *redis.Client) error {
//	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
//	rt := time.Unix(td.RtExpires, 0)
//	now := time.Now()
//
//	errAccess := client.Set(td.AccessUuid, userid, at.Sub(now)).Err()
//	if errAccess != nil {
//		return errAccess
//	}
//	errRefresh := client.Set(td.RefreshUuid, userid, rt.Sub(now)).Err()
//	if errRefresh != nil {
//		return errRefresh
//	}
//	return nil
//}
//
//func (h *TokenHelper)deleteAuth(givenUuid string, client *redis.Client) (int64, error) {
//	deleted, err := client.Del(givenUuid).Result()
//	if err != nil {
//		return 0, err
//	}
//	return deleted, nil
//}
//
//func deleteTokens(authD *AccessDetails, client *redis.Client) error {
//	//get the refresh uuid
//	refreshUuid := fmt.Sprintf("%s++%s", authD.AccessUuid, authD.UserID)
//	//delete access token
//	deletedAt, err := client.Del(authD.AccessUuid).Result()
//	if err != nil {
//		return err
//	}
//	//delete refresh token
//	deletedRt, err := client.Del(refreshUuid).Result()
//	if err != nil {
//		return err
//	}
//	//When the record is deleted, the return value is 1
//	if deletedAt != 1 || deletedRt != 1 {
//		return errors.New("something went wrong")
//	}
//	return nil
//}
