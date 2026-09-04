package access

// Тесты Б1 (TECH_DEBT.md): верификация подписи JWT.
// Покрытие: валидный токен keycloak (RS256), фейковая подпись, просроченный,
// malformed (alg=none / 2 сегмента), недоступный JWKS (fail-open и fail-closed),
// HS256 (sprob), интеграция RolesFromRequest.

import (
	"bytes"
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// signToken подписывает JWT (RS256/HS256) для тестов — без внешних зависимостей.
func signToken(t *testing.T, key *rsa.PrivateKey, secret, kid, alg string, claims map[string]interface{}) string {
	t.Helper()
	header := map[string]interface{}{"alg": alg, "typ": "JWT"}
	if kid != "" {
		header["kid"] = kid
	}
	hb, err := json.Marshal(header)
	if err != nil {
		t.Fatal(err)
	}
	pb, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	unsigned := base64.RawURLEncoding.EncodeToString(hb) + "." + base64.RawURLEncoding.EncodeToString(pb)

	var sig []byte
	switch alg {
	case "RS256":
		digest := sha256.Sum256([]byte(unsigned))
		sig, err = rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
		if err != nil {
			t.Fatal(err)
		}
	case "HS256":
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(unsigned))
		sig = mac.Sum(nil)
	default:
		t.Fatalf("unsupported alg %s", alg)
	}
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(sig)
}

// newJWKSServer поднимает httptest-сервер, отдающий JWKS с одним RSA-ключом.
func newJWKSServer(t *testing.T, kid string, pub *rsa.PublicKey) *httptest.Server {
	t.Helper()
	key := jwk{
		Kid: kid, Kty: "RSA", Alg: "RS256", Use: "sig",
		N: base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		E: base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(jwksResponse{Keys: []jwk{key}})
	}))
	t.Cleanup(srv.Close)
	return srv
}

func newVerifier(t *testing.T, jwksURL string) *TokenVerifier {
	t.Helper()
	return &TokenVerifier{jwksURL: jwksURL, client: &http.Client{Timeout: httpTimeout}}
}

func withDefaultVerifier(t *testing.T, v *TokenVerifier) {
	t.Helper()
	old := defaultVerifier
	defaultVerifier = v
	t.Cleanup(func() { defaultVerifier = old })
}

func TestVerifyValidKeycloakToken(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	srv := newJWKSServer(t, "kid1", &key.PublicKey)
	v := newVerifier(t, srv.URL)

	claims := map[string]interface{}{
		"sub":               "user-1",
		"preferred_username": "admin",
		"exp":               float64(time.Now().Add(time.Hour).Unix()),
		"realm_access":      map[string]interface{}{"roles": []interface{}{"R00_ADMIN"}},
	}
	tok := signToken(t, key, "", "kid1", "RS256", claims)

	got, err := v.Verify(tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if got["sub"] != "user-1" {
		t.Fatalf("sub = %v, want user-1", got["sub"])
	}
}

func TestVerifyFakeSignature(t *testing.T) {
	// Ключ JWKS и ключ подписи — разные: подделанный токен обязан отклоняться.
	realKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	fakeKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := newJWKSServer(t, "kid1", &realKey.PublicKey)
	v := newVerifier(t, srv.URL)

	claims := map[string]interface{}{
		"sub": "attacker",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
		"realm_access": map[string]interface{}{
			"roles": []interface{}{"R00_ADMIN"}, // подделываем админа
		},
	}
	tok := signToken(t, fakeKey, "", "kid1", "RS256", claims)

	if _, err := v.Verify(tok); err != ErrInvalidToken {
		t.Fatalf("err = %v, want ErrInvalidToken", err)
	}
}

func TestVerifyExpiredToken(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := newJWKSServer(t, "kid1", &key.PublicKey)
	v := newVerifier(t, srv.URL)

	claims := map[string]interface{}{
		"sub": "user-1",
		"exp": float64(time.Now().Add(-time.Hour).Unix()),
	}
	tok := signToken(t, key, "", "kid1", "RS256", claims)

	if _, err := v.Verify(tok); err != ErrInvalidToken {
		t.Fatalf("err = %v, want ErrInvalidToken", err)
	}

	// Б4: VerifyNoExp проверяет только подпись — истёкший токен проходит
	// (нужно для logout через sendBeacon с reason=token-expired).
	got, err := v.VerifyNoExp(tok)
	if err != nil {
		t.Fatalf("VerifyNoExp: %v", err)
	}
	if got["sub"] != "user-1" {
		t.Fatalf("sub = %v, want user-1", got["sub"])
	}
}

func TestVerifyNoExpFakeSignature(t *testing.T) {
	// Подпись фейковая — VerifyNoExp обязан отклонить (identity не доверяем).
	fakeKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	realKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := newJWKSServer(t, "kid1", &realKey.PublicKey)
	v := newVerifier(t, srv.URL)

	claims := map[string]interface{}{"sub": "attacker", "exp": float64(time.Now().Add(-time.Hour).Unix())}
	tok := signToken(t, fakeKey, "", "kid1", "RS256", claims)

	if _, err := v.VerifyNoExp(tok); err != ErrInvalidToken {
		t.Fatalf("err = %v, want ErrInvalidToken", err)
	}
}

func TestVerifyTokenNoExpPublic(t *testing.T) {
	// Публичная обёртка VerifyTokenNoExp возвращает UserCtx с identity из claims.
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := newJWKSServer(t, "kid1", &key.PublicKey)
	withDefaultVerifier(t, newVerifier(t, srv.URL))

	claims := map[string]interface{}{
		"sub":               "user-1",
		"preferred_username": "ivanov",
		"exp":               float64(time.Now().Add(-time.Hour).Unix()), // уже истёк
		"resource_access": map[string]interface{}{
			"map-app": map[string]interface{}{"roles": []interface{}{"r03_rem"}},
		},
	}
	tok := signToken(t, key, "", "kid1", "RS256", claims)

	uc, err := VerifyTokenNoExp(tok)
	if err != nil {
		t.Fatalf("VerifyTokenNoExp: %v", err)
	}
	if uc.UserID != "user-1" || uc.UserName != "ivanov" || len(uc.Roles) != 1 || uc.Roles[0] != "R03_REM" {
		t.Fatalf("identity: %+v", uc)
	}
	if !uc.TokenValid {
		t.Fatalf("TokenValid = false, want true")
	}
}

func TestVerifyMalformedTokens(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := newJWKSServer(t, "kid1", &key.PublicKey)
	v := newVerifier(t, srv.URL)

	cases := []string{
		"",                            // пусто
		"abc",                         // 1 сегмент
		"a.b",                         // 2 сегмента (alg=none без подписи — раньше проходил)
		"!!!.###.$$$",                 // не base64
		"eyJhbGciOiJub25lIn0.e30.",   // alg=none, пустая подпись
		"eyJhbGciOiJSUzI1NiJ9.e30.X", // битая подпись
		"eyJhbGciOiJSUzM4NCJ9.e30.X", // неподдерживаемый alg (RS384)
	}
	for _, tc := range cases {
		if _, err := v.Verify(tc); err != ErrInvalidToken {
			t.Fatalf("Verify(%q) err = %v, want ErrInvalidToken", tc, err)
		}
	}
}

func TestVerifyUnknownKid(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := newJWKSServer(t, "kid1", &key.PublicKey)
	v := newVerifier(t, srv.URL)

	claims := map[string]interface{}{"exp": float64(time.Now().Add(time.Hour).Unix())}
	tok := signToken(t, key, "", "kid-OTHER", "RS256", claims)

	if _, err := v.Verify(tok); err != ErrInvalidToken {
		t.Fatalf("err = %v, want ErrInvalidToken", err)
	}
}

func TestVerifyJWKSUnavailable(t *testing.T) {
	// Сервер закрыт — JWKS недоступен.
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	url := dead.URL
	dead.Close()

	v := newVerifier(t, url)
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	claims := map[string]interface{}{"exp": float64(time.Now().Add(time.Hour).Unix())}
	tok := signToken(t, key, "", "kid1", "RS256", claims)

	if _, err := v.Verify(tok); err != ErrVerifierUnavailable {
		t.Fatalf("err = %v, want ErrVerifierUnavailable", err)
	}
}

func TestVerifyStaleCacheOnRefreshFailure(t *testing.T) {
	// JWKS сначала работает, потом начинает падать: со старым кэшем проверка
	// подписи должна продолжаться (не уходить в fail-open без подписи).
	key, _ := rsa.GenerateKey(rand.Reader, 2048)

	var mu sync.Mutex
	fail := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if fail {
			http.Error(w, "keycloak down", http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(jwksResponse{Keys: []jwk{
			{Kid: "kid1", Kty: "RSA", Alg: "RS256", Use: "sig",
				N: base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
				E: base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.PublicKey.E)).Bytes())},
		}})
	}))
	t.Cleanup(srv.Close)
	v := newVerifier(t, srv.URL)

	claims := map[string]interface{}{"exp": float64(time.Now().Add(time.Hour).Unix())}
	tok := signToken(t, key, "", "kid1", "RS256", claims)

	// Первый запрос — кэш загружается, токен валиден.
	if _, err := v.Verify(tok); err != nil {
		t.Fatalf("initial verify: %v", err)
	}

	// JWKS «ломается», кэш устаревает.
	mu.Lock()
	fail = true
	mu.Unlock()
	v.mu.Lock()
	v.fetched = time.Now().Add(-jwksTTL - time.Minute)
	v.mu.Unlock()

	// Со старым кэшем токен всё ещё проверяется (подпись, не fail-open).
	if _, err := v.Verify(tok); err != nil {
		t.Fatalf("verify with stale cache: %v", err)
	}

	// После неудачного обновления повторная попытка планируется через jwksFailTTL
	// (а не через полный jwksTTL), чтобы ротация ключей keycloak не уходила в 401 надолго.
	v.mu.RLock()
	nextRetry := time.Until(v.fetched.Add(jwksTTL))
	v.mu.RUnlock()
	if nextRetry < jwksFailTTL-5*time.Second || nextRetry > jwksFailTTL+5*time.Second {
		t.Fatalf("next retry in %v, want ~%v", nextRetry, jwksFailTTL)
	}

	// Фейковый токен отклоняется и со старым кэшем.
	fakeKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	fakeTok := signToken(t, fakeKey, "", "kid1", "RS256", claims)
	if _, err := v.Verify(fakeTok); err != ErrInvalidToken {
		t.Fatalf("fake token with stale cache: err = %v, want ErrInvalidToken", err)
	}
}

func TestVerifyHS256(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	claims := map[string]interface{}{"exp": float64(time.Now().Add(time.Hour).Unix())}
	tok := signToken(t, key, "super-secret", "", "HS256", claims)

	t.Run("секрет задан — валиден", func(t *testing.T) {
		t.Setenv("TOKEN_SECRET", "super-secret")
		v := newVerifier(t, "http://unused")
		if _, err := v.Verify(tok); err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
	})
	t.Run("секрет не задан — reject (дыра закрыта)", func(t *testing.T) {
		t.Setenv("TOKEN_SECRET", "")
		v := newVerifier(t, "http://unused")
		if _, err := v.Verify(tok); err != ErrInvalidToken {
			t.Fatalf("err = %v, want ErrInvalidToken", err)
		}
	})
	t.Run("неверный секрет — reject", func(t *testing.T) {
		t.Setenv("TOKEN_SECRET", "wrong-secret")
		v := newVerifier(t, "http://unused")
		if _, err := v.Verify(tok); err != ErrInvalidToken {
			t.Fatalf("err = %v, want ErrInvalidToken", err)
		}
	})
}

// verifyRequestToken: fail-open / fail-closed / выключено.

func TestVerifyRequestTokenFailOpen(t *testing.T) {
	t.Setenv("JWT_VERIFY", "true")           // верификация включена
	t.Setenv("JWT_VERIFY_FAIL_OPEN", "true") // fail-open по умолчанию

	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	url := dead.URL
	dead.Close()
	withDefaultVerifier(t, newVerifier(t, url))

	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	claims := map[string]interface{}{"sub": "u1", "exp": float64(time.Now().Add(time.Hour).Unix())}
	tok := signToken(t, key, "", "kid1", "RS256", claims)

	got, status := verifyRequestToken(tok)
	if status != tokenStatusUnverified {
		t.Fatalf("status = %v, want tokenStatusUnverified", status)
	}
	if got["sub"] != "u1" {
		t.Fatalf("claims not parsed in fail-open: %v", got)
	}
}

func TestVerifyRequestTokenFailClosed(t *testing.T) {
	t.Setenv("JWT_VERIFY", "true")
	t.Setenv("JWT_VERIFY_FAIL_OPEN", "false") // fail-closed

	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	url := dead.URL
	dead.Close()
	withDefaultVerifier(t, newVerifier(t, url))

	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	tok := signToken(t, key, "", "kid1", "RS256", map[string]interface{}{"exp": float64(time.Now().Add(time.Hour).Unix())})

	if _, status := verifyRequestToken(tok); status != tokenStatusInvalid {
		t.Fatalf("status = %v, want tokenStatusInvalid", status)
	}
}

func TestVerifyRequestTokenDisabled(t *testing.T) {
	t.Setenv("JWT_VERIFY", "false") // аварийный рубильник

	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	claims := map[string]interface{}{"sub": "u1", "exp": float64(time.Now().Add(time.Hour).Unix())}
	tok := signToken(t, key, "", "kid1", "RS256", claims)

	got, status := verifyRequestToken(tok)
	if status != tokenStatusDisabled {
		t.Fatalf("status = %v, want tokenStatusDisabled", status)
	}
	if got["sub"] != "u1" {
		t.Fatalf("claims not parsed in disabled mode: %v", got)
	}
}

// Интеграция с RolesFromRequest.

func TestRolesFromRequestNoToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/buildings", nil)
	uc := RolesFromRequest(req.Context(), nil, req)
	if uc.TokenInvalid || uc.TokenValid || len(uc.Roles) != 0 {
		t.Fatalf("anonymous request: unexpected ctx %+v", uc)
	}
}

func TestRolesFromRequestInvalidToken(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	realKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := newJWKSServer(t, "kid1", &realKey.PublicKey)
	withDefaultVerifier(t, newVerifier(t, srv.URL))

	// Фейковый токен с ролью R00_ADMIN — как в описании Б1.
	claims := map[string]interface{}{
		"sub": "attacker",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
		"realm_access": map[string]interface{}{
			"roles": []interface{}{"R00_ADMIN"},
		},
	}
	tok := signToken(t, key, "", "kid1", "RS256", claims)

	req := httptest.NewRequest(http.MethodGet, "/api/buildings", nil)
	req.Header.Set("token", tok)
	uc := RolesFromRequest(req.Context(), nil, req)
	if !uc.TokenInvalid {
		t.Fatalf("TokenInvalid = false, want true")
	}
	if len(uc.Roles) != 0 || uc.UserID != "" {
		t.Fatalf("invalid token must not expose claims: %+v", uc)
	}
}

func TestRolesFromRequestValidToken(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := newJWKSServer(t, "kid1", &key.PublicKey)
	withDefaultVerifier(t, newVerifier(t, srv.URL))

	claims := map[string]interface{}{
		"sub":               "user-1",
		"preferred_username": "ivanov",
		"exp":               float64(time.Now().Add(time.Hour).Unix()),
		"realm_access": map[string]interface{}{
			"roles": []interface{}{"offline_access", "default-roles-rdkb"}, // служебные
		},
		"resource_access": map[string]interface{}{
			"map-app": map[string]interface{}{"roles": []interface{}{"r03_rem", "r03_rem"}},
			"hr-app":  map[string]interface{}{"roles": []interface{}{"HR_ADMIN"}}, // чужая роль
		},
	}
	tok := signToken(t, key, "", "kid1", "RS256", claims)

	req := httptest.NewRequest(http.MethodGet, "/api/buildings", nil)
	req.Header.Set("token", tok)
	uc := RolesFromRequest(req.Context(), nil, req)
	if !uc.TokenValid {
		t.Fatalf("TokenValid = false, want true")
	}
	if uc.UserID != "user-1" || uc.UserName != "ivanov" {
		t.Fatalf("identity: %+v", uc)
	}
	// Только роли map-app, нормализованы в верхний регистр, без дублей и служебных.
	if len(uc.Roles) != 1 || uc.Roles[0] != "R03_REM" {
		t.Fatalf("roles = %v, want [R03_REM]", uc.Roles)
	}
}

func TestRolesFromRequestAuthorizationHeader(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := newJWKSServer(t, "kid1", &key.PublicKey)
	withDefaultVerifier(t, newVerifier(t, srv.URL))

	claims := map[string]interface{}{
		"sub": "user-2",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
		"resource_access": map[string]interface{}{
			"map-app": map[string]interface{}{"roles": []interface{}{"r00_admin"}},
		},
	}
	tok := signToken(t, key, "", "kid1", "RS256", claims)

	req := httptest.NewRequest(http.MethodGet, "/api/buildings", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	uc := RolesFromRequest(req.Context(), nil, req)
	if !uc.TokenValid || len(uc.Roles) != 1 || uc.Roles[0] != "R00_ADMIN" {
		t.Fatalf("Bearer token: %+v", uc)
	}
}

// AccessControl: невалидный токен → 401 (таска 3 Б1).

func TestAccessControlInvalidToken401(t *testing.T) {
	fakeKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	realKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := newJWKSServer(t, "kid1", &realKey.PublicKey)
	withDefaultVerifier(t, newVerifier(t, srv.URL))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(NewMiddleware(nil).AccessControl())
	r.GET("/api/buildings", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	claims := map[string]interface{}{
		"sub": "attacker",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
		"realm_access": map[string]interface{}{
			"roles": []interface{}{"R00_ADMIN"},
		},
	}
	tok := signToken(t, fakeKey, "", "kid1", "RS256", claims)

	req := httptest.NewRequest(http.MethodGet, "/api/buildings", nil)
	req.Header.Set("token", tok)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code = %d, want 401", w.Code)
	}
}

func TestAccessControlNoTokenAnonymous(t *testing.T) {
	// Без токена запрос остаётся анонимным (datasync/upload'ы не ломаются).
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(NewMiddleware(nil).AccessControl())
	r.GET("/api/uploads/workers", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	r.GET("/api/buildings", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	req := httptest.NewRequest(http.MethodGet, "/api/uploads/workers", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("uploads без токена: code = %d, want 200", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/buildings", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("entity без токена: code = %d, want 200 (анонимный доступ до enforcement)", w.Code)
	}
}

func TestAccessControlValidTokenPasses(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := newJWKSServer(t, "kid1", &key.PublicKey)
	withDefaultVerifier(t, newVerifier(t, srv.URL))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(NewMiddleware(nil).AccessControl())
	// Не-сущностный роут: маскирование по матрице не применяется → DB не нужна.
	r.GET("/api/uploads/workers", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	// Валидный токен с ролью FM проходит (enforcement выключен).
	claims := map[string]interface{}{
		"sub": "user-1",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
		"resource_access": map[string]interface{}{
			"map-app": map[string]interface{}{"roles": []interface{}{"r03_rem"}},
		},
	}
	tok := signToken(t, key, "", "kid1", "RS256", claims)

	req := httptest.NewRequest(http.MethodGet, "/api/uploads/workers", nil)
	req.Header.Set("token", tok)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", w.Code)
	}
}

// Б3: режимы enforcement — monitor (WOULD-403) и true (403).

// токен без ролей и без user_id (фолбэк user_roles не трогает БД).
func noRoleToken(t *testing.T, key *rsa.PrivateKey, kid string) string {
	t.Helper()
	claims := map[string]interface{}{
		"exp":  float64(time.Now().Add(time.Hour).Unix()),
		"jti":  "no-identity",
	}
	return signToken(t, key, "", kid, "RS256", claims)
}

func TestAccessControlEnforceReadForbidden(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := newJWKSServer(t, "kid1", &key.PublicKey)
	withDefaultVerifier(t, newVerifier(t, srv.URL))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(NewMiddleware(nil).AccessControl())
	r.GET("/api/buildings", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	req := httptest.NewRequest(http.MethodGet, "/api/buildings", nil)
	req.Header.Set("token", noRoleToken(t, key, "kid1"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// По умолчанию (ACCESS_ENFORCE пуст) — 200 (обратная совместимость).
	if w.Code != http.StatusOK {
		t.Fatalf("default: code = %d, want 200", w.Code)
	}
}

func TestAccessControlMonitorLogsWould403(t *testing.T) {
	t.Setenv("ACCESS_ENFORCE", "monitor")
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := newJWKSServer(t, "kid1", &key.PublicKey)
	withDefaultVerifier(t, newVerifier(t, srv.URL))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(NewMiddleware(nil).AccessControl())
	r.GET("/api/buildings", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	req := httptest.NewRequest(http.MethodGet, "/api/buildings", nil)
	req.Header.Set("token", noRoleToken(t, key, "kid1"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// monitor: запрос НЕ блокируется (200), а WOULD-403 пишется в лог.
	if w.Code != http.StatusOK {
		t.Fatalf("monitor: code = %d, want 200 (запрос не блокируется)", w.Code)
	}
}

func TestAccessControlEnforceTrueForbidden(t *testing.T) {
	t.Setenv("ACCESS_ENFORCE", "true")
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := newJWKSServer(t, "kid1", &key.PublicKey)
	withDefaultVerifier(t, newVerifier(t, srv.URL))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(NewMiddleware(nil).AccessControl())
	r.GET("/api/buildings", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	req := httptest.NewRequest(http.MethodGet, "/api/buildings", nil)
	req.Header.Set("token", noRoleToken(t, key, "kid1"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// true: чтение без ролей → 403.
	if w.Code != http.StatusForbidden {
		t.Fatalf("enforce=true: code = %d, want 403", w.Code)
	}
}

// ===== Б3: сценарии записи (PUT/POST) в режимах enforce/monitor =====

// roleToken - токен с клиентскими ролями keycloak (resource_access.map-app.roles).
func roleToken(t *testing.T, key *rsa.PrivateKey, kid string, roles ...string) string {
	t.Helper()
	roleList := make([]interface{}, len(roles))
	for i, r := range roles {
		roleList[i] = r
	}
	claims := map[string]interface{}{
		"exp": float64(time.Now().Add(time.Hour).Unix()),
		"sub": "11111111-1111-1111-1111-111111111111",
		"resource_access": map[string]interface{}{
			"map-app": map[string]interface{}{"roles": roleList},
		},
	}
	return signToken(t, key, "", kid, "RS256", claims)
}

// seedMatrix загружает матрицу доступа в store без БД (тесты).
func seedMatrix(m *Middleware, rows map[string]map[string]map[string]string) {
	m.matrix.mu.Lock()
	defer m.matrix.mu.Unlock()
	m.matrix.matrix = rows
	m.matrix.loaded = time.Now()
}

// withFakeOldRow подменяет чтение текущей строки БД (diff-проверка PUT).
func withFakeOldRow(t *testing.T, row map[string]interface{}) {
	t.Helper()
	old := fetchOldRowFn
	fetchOldRowFn = func(_ *Middleware, _ context.Context, _ *Entity, _ string) map[string]interface{} {
		return row
	}
	t.Cleanup(func() { fetchOldRowFn = old })
}

// formRequest - запрос записи в формате sprof: multipart с JSON в поле `form`.
func formRequest(t *testing.T, method, url string, payload map[string]interface{}) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := mw.WriteField("form", string(b)); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(method, url, &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("token", roleToken(t, mustRSAKey(t), "kid1", "r01_hoz"))
	return req
}

// mustRSAKey - общий ключ для подписи тестовых токенов.
var testKey *rsa.PrivateKey

func mustRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	if testKey == nil {
		k, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatal(err)
		}
		testKey = k
	}
	return testKey
}

func newEnforceRouter(t *testing.T, rows map[string]map[string]map[string]string) *gin.Engine {
	t.Helper()
	key := mustRSAKey(t)
	srv := newJWKSServer(t, "kid1", &key.PublicKey)
	withDefaultVerifier(t, newVerifier(t, srv.URL))
	gin.SetMode(gin.TestMode)
	mw := NewMiddleware(nil)
	seedMatrix(mw, rows)
	r := gin.New()
	r.Use(mw.AccessControl())
	r.PUT("/api/rooms/:id", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	r.POST("/api/rooms", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	r.PUT("/api/room-engineerings/:id", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	return r
}

// Матрица для тестов: у R01_HOZ есть W на rooms.name; на остальные поля — R/нет.
var testMatrix = map[string]map[string]map[string]string{
	"rooms": {
		"name":     {"R01_HOZ": "W", "R00_ADMIN": "W"},
		"area":     {"R00_ADMIN": "W"},
		"actualName": {"R01_HOZ": "W", "R00_ADMIN": "W"},
	},
	"room-engineerings": {
		"electricity": {"R02_EXPL": "W", "R00_ADMIN": "W"},
	},
}

// PUT с полным объектом: не-W поле не меняется → запись разрешена (клиент шлёт весь объект).
func TestAccessControlEnforcePutUnchangedNonWFieldAllowed(t *testing.T) {
	t.Setenv("ACCESS_ENFORCE", "true")
	withFakeOldRow(t, map[string]interface{}{"name": "Кабинет 101", "area": float64(42)})
	r := newEnforceRouter(t, testMatrix)

	payload := map[string]interface{}{"id": "123", "name": "Кабинет 101", "area": float64(42)}
	req := formRequest(t, http.MethodPut, "/api/rooms/123", payload)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200 (не-W поля не меняются)", w.Code)
	}
}

// PUT с изменённым не-W полем → 403.
func TestAccessControlEnforcePutChangedDenied403(t *testing.T) {
	t.Setenv("ACCESS_ENFORCE", "true")
	withFakeOldRow(t, map[string]interface{}{"name": "Кабинет 101", "area": float64(42)})
	r := newEnforceRouter(t, testMatrix)

	// area — без W у R01_HOZ, значение меняется → 403
	payload := map[string]interface{}{"id": "123", "name": "Кабинет 101", "area": float64(99)}
	req := formRequest(t, http.MethodPut, "/api/rooms/123", payload)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("code = %d, want 403 (изменено не-W поле)", w.Code)
	}
}

// PUT с изменённым W-полем → 200.
func TestAccessControlEnforcePutChangedWField200(t *testing.T) {
	t.Setenv("ACCESS_ENFORCE", "true")
	withFakeOldRow(t, map[string]interface{}{"name": "Кабинет 101", "area": float64(42)})
	r := newEnforceRouter(t, testMatrix)

	payload := map[string]interface{}{"id": "123", "name": "Кабинет 102", "area": float64(42)}
	req := formRequest(t, http.MethodPut, "/api/rooms/123", payload)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200 (изменено W-поле)", w.Code)
	}
}

// POST (create) с не-W полем → 403: при создании проверяются все поля.
func TestAccessControlEnforcePostDenied403(t *testing.T) {
	t.Setenv("ACCESS_ENFORCE", "true")
	r := newEnforceRouter(t, testMatrix)

	payload := map[string]interface{}{"name": "Новая комната", "area": float64(10)}
	req := formRequest(t, http.MethodPost, "/api/rooms", payload)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("code = %d, want 403 (create с не-W полем)", w.Code)
	}
}

// Entity-level сущность: без W на сущности → 403 (даже если поля не меняются).
func TestAccessControlEnforcePutEntityLevelDenied403(t *testing.T) {
	t.Setenv("ACCESS_ENFORCE", "true")
	r := newEnforceRouter(t, testMatrix)

	payload := map[string]interface{}{"id": "123", "electricity": "есть"}
	req := formRequest(t, http.MethodPut, "/api/room-engineerings/123", payload)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("code = %d, want 403 (R01_HOZ без W на room-engineerings)", w.Code)
	}
}

// monitor: изменение не-W поля не блокируется (200), пишется WOULD-403.
func TestAccessControlMonitorPutChangedDenied200(t *testing.T) {
	t.Setenv("ACCESS_ENFORCE", "monitor")
	withFakeOldRow(t, map[string]interface{}{"name": "Кабинет 101", "area": float64(42)})
	r := newEnforceRouter(t, testMatrix)

	payload := map[string]interface{}{"id": "123", "name": "Кабинет 101", "area": float64(99)}
	req := formRequest(t, http.MethodPut, "/api/rooms/123", payload)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200 (monitor не блокирует)", w.Code)
	}
}

// fieldChanged: сравнение значений (строки, числа, время в разных форматах).
func TestFieldChanged(t *testing.T) {
	row := map[string]interface{}{
		"name":       "Кабинет 101",
		"area":       float64(42),
		"worker_id":  "abc",
		"sout_date":  "2026-08-18 12:00:00+00:00",
	}
	cases := []struct {
		name  string
		field string
		val   interface{}
		want  bool
	}{
		{"строка не изменилась", "name", "Кабинет 101", false},
		{"строка изменилась", "name", "Другое", true},
		{"число не изменилось", "area", float64(42), false},
		{"число изменилось", "area", float64(43), true},
		{"связь _id не изменилась", "worker", "abc", false},
		{"связь _id изменилась", "worker", "xyz", true},
		{"время ISO == БД", "soutDate", "2026-08-18T12:00:00Z", false},
		{"время изменилось", "soutDate", "2026-08-19T12:00:00Z", true},
		{"нет в строке БД — изменено", "unknown", "x", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := fieldChanged(row, tc.field, tc.val); got != tc.want {
				t.Fatalf("fieldChanged(%q, %v) = %v, want %v", tc.field, tc.val, got, tc.want)
			}
		})
	}
}
