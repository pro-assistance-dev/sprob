package access

// Б1 (TECH_DEBT.md): верификация подписи JWT на бэкенде.
//
// До Б1 `decodeJWTClaims` разбирал payload без проверки подписи — любой запрос
// с заголовком `token: <base64>.fake` проходил как R00_ADMIN (прод-nginx
// проксирует /api/ напрямую, без oauth2-proxy auth_request).
//
// Теперь токены проверяются:
//   - RS256 — по JWKS keycloak (GET {JWT_JWKS_URL}/realms/rdkb/protocol/openid-connect/certs),
//     ключи кэшируются (TTL 10 мин), при недоступности — поведение зависит от
//     JWT_VERIFY_FAIL_OPEN;
//   - HS256 — по TOKEN_SECRET (sprob-токены); если секрет не задан — reject
//     (иначе HS256 с пустым секретом остаётся дырой).
//
// Конфигурация (env):
//   - JWT_VERIFY            — "false" отключает всю верификацию (аварийный рубильник; по умолчанию true);
//   - JWT_JWKS_URL          — базовый URL keycloak (по умолчанию http://rdkb-keycloak:80);
//   - JWT_VERIFY_FAIL_OPEN  — "false" = fail-closed (при недоступности JWKS токен
//     считается невалидным); по умолчанию true = fail-open с WARNING в лог,
//     чтобы падение keycloak не ломало прод.
//
// Поведение безопасно для прода:
//   - запросы БЕЗ токена остаются анонимными (datasync hr→map, uploads и т.п.);
//   - токен с невалидной подписью → 401 (middleware AccessControl);
//   - keycloak недоступен → fail-open (по умолчанию) + WARNING в лог.

import (
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Ошибки верификации токена.
var (
	// ErrInvalidToken — подпись невалидна, токен просрочен, malformed или
	// использует неподдерживаемый алгоритм.
	ErrInvalidToken = errors.New("access: invalid token")
	// ErrVerifierUnavailable — JWKS недоступен (keycloak не отвечает / нет ключей).
	// Что с этим делать — решает JWT_VERIFY_FAIL_OPEN.
	ErrVerifierUnavailable = errors.New("access: jwks unavailable")
)

const (
	defaultKeycloakURL = "http://rdkb-keycloak:80"
	jwksPath           = "/realms/rdkb/protocol/openid-connect/certs"
	jwksTTL            = 10 * time.Minute // свежесть кэша ключей
	jwksFailTTL        = 60 * time.Second // повторная попытка после ошибки
	httpTimeout        = 3 * time.Second
	clockSkew          = 30 * time.Second // допуск на рассинхрон часов для exp
)

// jwk — одна запись JWKS (поле RSA).
type jwk struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwksResponse struct {
	Keys []jwk `json:"keys"`
}

// TokenVerifier — верификатор JWT: кэш JWKS-ключей keycloak + HS256-секрет.
type TokenVerifier struct {
	mu      sync.RWMutex
	keys    map[string]*rsa.PublicKey
	fetched time.Time
	lastErr error

	jwksURL string
	client  *http.Client
}

// defaultVerifier — глобальный верификатор. URL читается из env при первом
// использовании (JWT_JWKS_URL), чтобы dev/prod могли отличаться без пересборки.
var defaultVerifier = &TokenVerifier{client: &http.Client{Timeout: httpTimeout}}

func (v *TokenVerifier) jwksEndpoint() string {
	if v.jwksURL != "" {
		return v.jwksURL
	}
	base := os.Getenv("JWT_JWKS_URL")
	if base == "" {
		base = defaultKeycloakURL
	}
	return strings.TrimRight(base, "/") + jwksPath
}

// Verify проверяет подпись JWT и возвращает claims.
// Ошибки: ErrInvalidToken | ErrVerifierUnavailable.
func (v *TokenVerifier) Verify(raw string) (map[string]interface{}, error) {
	parts := strings.Split(raw, ".")
	// Токен обязан быть полным: header.payload.signature. Двухсегментный
	// `header.payload` (alg=none) раньше проходил — теперь reject.
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, ErrInvalidToken
	}

	switch header.Alg {
	case "RS256":
		return v.verifyRS256(parts, header.Kid, true)
	case "HS256":
		return verifyHS256(parts, true)
	default:
		// alg=none, RS384/512 и всё прочее — не поддерживаем.
		return nil, ErrInvalidToken
	}
}

// VerifyNoExp проверяет подпись JWT БЕЗ проверки срока действия (exp).
// Нужен для Б4: logout через navigator.sendBeacon при закрытии вкладки —
// к этому моменту access token может быть уже истёкшим (reason=token-expired),
// но подпись обязана быть валидной (иначе identity не доверяем).
func (v *TokenVerifier) VerifyNoExp(raw string) (map[string]interface{}, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, ErrInvalidToken
	}
	switch header.Alg {
	case "RS256":
		return v.verifyRS256(parts, header.Kid, false)
	case "HS256":
		return verifyHS256(parts, false)
	default:
		return nil, ErrInvalidToken
	}
}

// verifyRS256 проверяет подпись RSA-PKCS1v15/SHA-256 по ключу из JWKS (по kid).
func (v *TokenVerifier) verifyRS256(parts []string, kid string, checkExpiry bool) (map[string]interface{}, error) {
	keys, err := v.getKeys()
	if err != nil && len(keys) == 0 {
		// JWKS недоступен и кэша (свежего или старого) нет вовсе.
		return nil, ErrVerifierUnavailable
	}
	// При ошибке обновления со старым кэшем продолжаем проверять старыми ключами
	// (warn-лог пишется в getKeys) — это безопаснее, чем fail-open без подписи.
	key, ok := keys[kid]
	if !ok {
		return nil, ErrInvalidToken
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, ErrInvalidToken
	}
	signed := parts[0] + "." + parts[1]
	digest := sha256.Sum256([]byte(signed))
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], sig); err != nil {
		return nil, ErrInvalidToken
	}
	claims := decodeJWTClaims(strings.Join(parts, "."))
	if checkExpiry && !checkExp(claims) {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// verifyHS256 проверяет подпись HMAC-SHA256 по TOKEN_SECRET (sprob-токены).
// Если секрет не задан — reject: HS256 с пустым секретом подделывается кем угодно.
func verifyHS256(parts []string, checkExpiry bool) (map[string]interface{}, error) {
	secret := os.Getenv("TOKEN_SECRET")
	if secret == "" {
		return nil, ErrInvalidToken
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, ErrInvalidToken
	}
	signed := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signed))
	if !hmac.Equal(mac.Sum(nil), sig) {
		return nil, ErrInvalidToken
	}
	claims := decodeJWTClaims(strings.Join(parts, "."))
	if checkExpiry && !checkExp(claims) {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// checkExp отклоняет токены без exp или с истёкшим сроком (с допуском clockSkew).
func checkExp(claims map[string]interface{}) bool {
	exp, ok := claims["exp"].(float64)
	if !ok {
		return false
	}
	return time.Now().Before(time.Unix(int64(exp), 0).Add(clockSkew))
}

// getKeys возвращает кэш RSA-ключей (kid → ключ), при необходимости обновляя его.
// Возвращает ErrVerifierUnavailable, если JWKS недоступен и свежего кэша нет.
func (v *TokenVerifier) getKeys() (map[string]*rsa.PublicKey, error) {
	v.mu.RLock()
	keys, fetched, lastErr := v.keys, v.fetched, v.lastErr
	have := len(keys) > 0
	v.mu.RUnlock()

	now := time.Now()
	// Свежий кэш — отдаём без запросов (быстрый путь каждого запроса).
	if have && now.Sub(fetched) < jwksTTL {
		return keys, nil
	}
	// Нет кэша и недавняя ошибка — не долбим keycloak.
	if !have && lastErr != nil && now.Sub(fetched) < jwksFailTTL {
		return nil, lastErr
	}

	// Обновление (упрощённый singleflight: повторная проверка под блокировкой).
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.keys != nil && time.Since(v.fetched) < jwksTTL {
		return v.keys, nil
	}
	if v.keys == nil && v.lastErr != nil && time.Since(v.fetched) < jwksFailTTL {
		return nil, v.lastErr
	}

	newKeys, err := v.fetch()
	if err != nil {
		v.lastErr = err
		if len(v.keys) > 0 {
			// Есть старый кэш — продолжаем проверять им, но повторяем обновление
			// через jwksFailTTL (не через полный jwksTTL): если keycloak ротировал
			// ключи, новые токены (новый kid) не должны уходить в 401 надолго.
			v.fetched = time.Now().Add(-jwksTTL + jwksFailTTL)
			log.Printf("[access][jwt] WARNING: обновление JWKS не удалось (%v); используем кэш от %s, повтор через %s",
				err, time.Now().Add(-jwksTTL).Format(time.RFC3339), jwksFailTTL)
		} else {
			v.fetched = time.Now()
			log.Printf("[access][jwt] ERROR: JWKS недоступен и кэша нет (%v)", err)
		}
		return v.keys, err // v.keys может быть nil (первый запуск)
	}
	v.fetched = time.Now()
	v.keys, v.lastErr = newKeys, nil
	return newKeys, nil
}

// fetch загружает JWKS и декодирует RSA-ключи.
func (v *TokenVerifier) fetch() (map[string]*rsa.PublicKey, error) {
	resp, err := v.client.Get(v.jwksEndpoint())
	if err != nil {
		return nil, fmt.Errorf("jwks fetch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jwks fetch: status %d", resp.StatusCode)
	}
	var jr jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jr); err != nil {
		return nil, fmt.Errorf("jwks decode: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey, len(jr.Keys))
	for _, k := range jr.Keys {
		if k.Kty != "" && k.Kty != "RSA" {
			continue
		}
		if k.Use != "" && k.Use != "sig" {
			continue
		}
		pub, err := k.publicKey()
		if err != nil {
			continue // пропускаем битые ключи, остальные используем
		}
		keys[k.Kid] = pub
	}
	if len(keys) == 0 {
		return nil, errors.New("jwks: no usable RSA keys")
	}
	return keys, nil
}

// publicKey собирает *rsa.PublicKey из base64url n/e.
func (k jwk) publicKey() (*rsa.PublicKey, error) {
	nb, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, err
	}
	eb, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, err
	}
	e := 0
	for _, b := range eb {
		e = e<<8 | int(b)
	}
	if e == 0 {
		return nil, errors.New("jwks: zero exponent")
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nb), E: e}, nil
}

// ============================================================
// Публичный API для RolesFromRequest
// ============================================================

// tokenStatus — результат проверки токена.
type tokenStatus int

const (
	// tokenStatusValid — подпись проверена и валидна (claims можно доверять).
	tokenStatusValid tokenStatus = iota
	// tokenStatusInvalid — токен есть, но подпись/срок не прошли проверку.
	tokenStatusInvalid
	// tokenStatusUnverified — JWKS недоступен, включён fail-open: claims разобраны
	// БЕЗ проверки подписи (поведение до Б1), в лог записан WARNING.
	tokenStatusUnverified
	// tokenStatusDisabled — JWT_VERIFY=false: верификация выключена (аварийно).
	tokenStatusDisabled
	// tokenStatusNone — токена в запросе нет (аноним).
	tokenStatusNone
)

// verifyJWTEnabled — включена ли верификация (JWT_VERIFY != "false").
func verifyJWTEnabled() bool {
	return os.Getenv("JWT_VERIFY") != "false"
}

// jwtFailOpen — fail-open при недоступности JWKS (JWT_VERIFY_FAIL_OPEN != "false").
func jwtFailOpen() bool {
	return os.Getenv("JWT_VERIFY_FAIL_OPEN") != "false"
}

// verifyRequestToken проверяет токен из запроса и возвращает claims + статус.
func verifyRequestToken(raw string) (map[string]interface{}, tokenStatus) {
	if raw == "" {
		return nil, tokenStatusNone
	}
	if !verifyJWTEnabled() {
		// Аварийный режим: поведение до Б1 (без проверки подписи).
		return decodeJWTClaims(raw), tokenStatusDisabled
	}

	claims, err := defaultVerifier.Verify(raw)
	switch {
	case err == nil:
		return claims, tokenStatusValid
	case errors.Is(err, ErrInvalidToken):
		return nil, tokenStatusInvalid
	case errors.Is(err, ErrVerifierUnavailable):
		if jwtFailOpen() {
			log.Printf("[access][jwt] WARNING: JWKS недоступен (%v); токен НЕ верифицирован (fail-open). "+
				"Проверьте keycloak и JWT_JWKS_URL, либо JWT_VERIFY_FAIL_OPEN=false", err)
			return decodeJWTClaims(raw), tokenStatusUnverified
		}
		return nil, tokenStatusInvalid
	default:
		return nil, tokenStatusInvalid
	}
}

// VerifyTokenNoExp проверяет подпись токена БЕЗ проверки срока действия (exp)
// и возвращает identity из claims. Используется в Б4 для logout через
// navigator.sendBeacon (заголовков нет, токен мог уже истечь — reason=token-expired).
// Подпись обязана быть валидной, иначе identity не доверяем.
func VerifyTokenNoExp(raw string) (UserCtx, error) {
	claims, err := defaultVerifier.VerifyNoExp(raw)
	if err != nil {
		return UserCtx{}, err
	}
	uc := UserCtx{}
	fillUserCtx(&uc, claims)
	uc.TokenValid = true
	return uc, nil
}
