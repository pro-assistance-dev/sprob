package tokenHelper

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
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

type TokenDetails struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

type JWTClaimsSetter interface {
	SetJWTClaimsMap(claims map[string]interface{})
}

func (h *TokenHelper) GetSigned(claims jwt.MapClaims) (string, error) {
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return rt.SignedString([]byte(h.TokenSecret))
}

func (item *TokenDetails) SetAccessTokenClaims(claims jwt.MapClaims, exp int) {
	claims["authorized"] = true
	claims["access_uuid"] = uuid.NewString()
	claims["exp"] = time.Now().Add(time.Minute * time.Duration(exp)).Unix()
}

func (item *TokenDetails) SetRefreshTokenClaims(claims jwt.MapClaims, exp int) {
	claims["refresh_uuid"] = uuid.NewString()
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(exp)).Unix()
}

func (h *TokenHelper) CreateToken(claimsSetter JWTClaimsSetter) (td *TokenDetails, err error) {
	atClaims := jwt.MapClaims{}
	td.SetAccessTokenClaims(atClaims, h.TokenAccessMinutes)
	claimsSetter.SetJWTClaimsMap(atClaims)
	td.AccessToken, err = h.GetSigned(atClaims)
	if err != nil {
		return nil, err
	}

	rtClaims := jwt.MapClaims{}
	td.SetRefreshTokenClaims(rtClaims, h.TokenRefreshHours)
	td.RefreshToken, err = h.GetSigned(rtClaims)
	if err != nil {
		return nil, err
	}
	return td, nil
}
func (h *TokenHelper) RefreshToken(refreshToken string, claimsSetter JWTClaimsSetter) (*TokenDetails, error) {
	token, err := h.VerifyToken(refreshToken)
	if err != nil || !token.Valid {
		return nil, err
	}
	return h.CreateToken(claimsSetter)
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

func (h *TokenHelper) extractTokenMetadata(r *http.Request, claim string) (string, error) {
	token, err := h.VerifyToken(h.extractToken(r))
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", err
	}
	res, ok := claims[claim].(string)
	if !ok {
		return "", errors.New("claim not found")
	}
	return res, nil
}

func (h *TokenHelper) extractToken(r *http.Request) string {
	bearToken := r.Header.Get("token")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return bearToken
}
