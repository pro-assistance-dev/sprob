package tokenHelper

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pro-assistance/pro-assister/config"
)

type TokenHelper struct {
	TokenSecret        string
	TokenAccessMinutes int
	TokenRefreshHours  int
}

func NewTokenHelper(conf config.Token) *TokenHelper {
	return &TokenHelper{TokenSecret: conf.TokenSecret, TokenAccessMinutes: conf.TokenAccessMinutes, TokenRefreshHours: conf.TokenRefreshHours}
}

type AccessDetails struct {
	AccessUuid   string
	UserID       string
	UserDomainID string
	UserRole     string
	UserRoleID   string
}

type TokenDetails struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

func (h *TokenHelper) CreateToken(userID string, userRole string, userRoleID string, userDomainID string) (*TokenDetails, error) {
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * time.Duration(h.TokenAccessMinutes)).Unix()
	td.AccessUuid = uuid.NewString()

	td.RtExpires = time.Now().Add(time.Hour * time.Duration(h.TokenRefreshHours)).Unix()
	td.RefreshUuid = td.AccessUuid + "++" + userID

	var err error
	//Creating Access Token
	os.Setenv("ACCESS_SECRET", h.TokenSecret) //this should be in an env file
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["user_id"] = userID
	atClaims["user_role"] = userRole
	atClaims["user_domain_id"] = userDomainID
	atClaims["user_role_id"] = userRoleID
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
	rtClaims["user_role"] = userRole
	rtClaims["user_role_id"] = userRoleID
	rtClaims["user_domain_id"] = userDomainID
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
	if err != nil || !token.Valid {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, err
	}
	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, err
	}
	userRole, ok := claims["user_role"].(string)
	if !ok {
		return nil, err
	}
	userRoleID, ok := claims["user_role_id"].(string)
	if !ok {
		return nil, err
	}
	userDomainID, ok := claims["user_domain_id"].(string)
	if !ok {
		return nil, err
	}
	return h.CreateToken(userID, userRole, userRoleID, userDomainID)
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
	if !ok || !token.Valid {
		return nil, err
	}
	accessUuid, ok := claims["access_uuid"].(string)
	if !ok {
		return nil, err
	}
	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, err
	}
	userRole, ok := claims["user_role"].(string)
	if !ok {
		return nil, err
	}
	userRoleID, ok := claims["user_role_id"].(string)
	if !ok {
		return nil, err
	}
	userDomainID, ok := claims["user_domain_id"].(string)
	if !ok {
		return nil, err
	}
	return &AccessDetails{
		AccessUuid:   accessUuid,
		UserID:       userID,
		UserRole:     userRole,
		UserRoleID:   userRoleID,
		UserDomainID: userDomainID,
	}, nil

}

func (h *TokenHelper) extractToken(r *http.Request) string {
	bearToken := r.Header.Get("token")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return bearToken
}

func (h *TokenHelper) GetUserID(c *gin.Context) (*uuid.UUID, error) {
	accessDetails, err := h.extractTokenMetadata(c.Request)
	if err != nil {
		return nil, err
	}
	uuidFromString, err := uuid.Parse(accessDetails.UserID)
	return &uuidFromString, err
}

func (h *TokenHelper) GetUserRole(c *gin.Context) (string, error) {
	accessDetails, err := h.extractTokenMetadata(c.Request)
	if err != nil {
		return "", err
	}
	return accessDetails.UserRole, err
}

func (h *TokenHelper) GetUserDomainID(c *gin.Context) (string, error) {
	accessDetails, err := h.extractTokenMetadata(c.Request)
	if err != nil {
		return "", err
	}
	return accessDetails.UserDomainID, err
}

func (h *TokenHelper) GetAccessDetail(c *gin.Context) (*AccessDetails, error) {
	accessDetails, err := h.extractTokenMetadata(c.Request)
	if err != nil {
		return nil, err
	}
	return accessDetails, err
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
