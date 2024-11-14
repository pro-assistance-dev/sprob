package token

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/pro-assistance-dev/sprob/config"
)

type Token struct {
	TokenSecret        string
	TokenAccessMinutes int
	TokenRefreshHours  int
}

func NewToken(conf config.Token) *Token {
	return &Token{TokenSecret: conf.TokenSecret, TokenAccessMinutes: conf.TokenAccessMinutes, TokenRefreshHours: conf.TokenRefreshHours}
}

type Details struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	// AccessUUID string `json:"accessUUID"`
	// RefreshUUID string `json:"refreshUuid"`
	AtExpires int64
	RtExpires int64
}

type JWTClaimsSetter interface {
	SetJWTClaimsMap(claims map[string]interface{})
}

func (h *Token) getSigned(claims jwt.MapClaims) (string, error) {
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return rt.SignedString([]byte(h.TokenSecret))
}

func (item *Details) setAccessTokenClaims(claims jwt.MapClaims, exp int) {
	claims["authorized"] = true
	claims["access_uuid"] = uuid.NewString()
	claims["exp"] = time.Now().Add(time.Minute * time.Duration(exp)).Unix()
}

func (item *Details) setRefreshTokenClaims(claims jwt.MapClaims, exp int) {
	claims["refresh_uuid"] = uuid.NewString()
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(exp)).Unix()
}

func (h *Token) CreateToken(claimsSetter JWTClaimsSetter) (td *Details, err error) {
	td = &Details{}
	atClaims := jwt.MapClaims{}
	td.setAccessTokenClaims(atClaims, h.TokenAccessMinutes)
	claimsSetter.SetJWTClaimsMap(atClaims)
	td.AccessToken, err = h.getSigned(atClaims)
	if err != nil {
		return nil, err
	}

	rtClaims := jwt.MapClaims{}
	td.setRefreshTokenClaims(rtClaims, h.TokenRefreshHours)
	td.RefreshToken, err = h.getSigned(rtClaims)
	if err != nil {
		return nil, err
	}
	return td, nil
}

func (h *Token) RefreshToken(refreshToken string, claimsSetter JWTClaimsSetter) (*Details, error) {
	token, err := h.verifyToken(refreshToken)
	if err != nil || !token.Valid {
		return nil, err
	}
	return h.CreateToken(claimsSetter)
}

func (h *Token) verifyToken(tokenString string) (*jwt.Token, error) {
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

func (h *Token) ExtractTokenMetadata(r *http.Request, claim fmt.Stringer) (string, error) {
	token, err := h.verifyToken(h.extractToken(r))
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", err
	}
	res, ok := claims[claim.String()].(string)
	if !ok {
		return "", errors.New("claim not found")
	}
	return res, nil
}

func (h *Token) extractToken(r *http.Request) string {
	bearToken := r.Header.Get("token")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return bearToken
}
