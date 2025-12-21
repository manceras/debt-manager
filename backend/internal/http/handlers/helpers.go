package handlers

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

type Argon2Params struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
	SaltLen uint32
}

var defalutA2 = Argon2Params{
	Time:    2,
	Memory:  64 * 1024,
	Threads: 1,
	KeyLen:  32,
	SaltLen: 16,
}

func HashPassword(password string) (string, error) {
	salt := make([]byte, defalutA2.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, defalutA2.Time, defalutA2.Memory, defalutA2.Threads, defalutA2.KeyLen)
	return base64.RawStdEncoding.EncodeToString(salt) + "$" + base64.RawStdEncoding.EncodeToString(hash), nil

}

func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.SplitN(encodedHash, "$", 2)
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid hash format")
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false, err
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false, err
	}
	computedHash := argon2.IDKey([]byte(password), salt, defalutA2.Time, defalutA2.Memory, defalutA2.Threads, defalutA2.KeyLen)
	return subtle.ConstantTimeCompare(hash, computedHash) == 1, nil
}

func setCookie(w http.ResponseWriter, name, value string, ttl time.Duration, path ...string) {
	p := "/"
	if len(path) > 0 {
		p = path[0]
	}
  http.SetCookie(w, &http.Cookie{
    Name:     name,
    Value:    value,
    HttpOnly: true,
    Secure:   true,
    SameSite: http.SameSiteLaxMode,
    Expires:  time.Now().Add(ttl),
		Path: 		p,
  })
}

func clearCookie(w http.ResponseWriter, name string, path ...string) {
	setCookie(w, name, "", -time.Hour, path...)
}

func GetPathParams(path string, paramNames ...string) map[string]string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < len(paramNames) {
		return nil
	}
	params := make(map[string]string)
	for i, name := range paramNames {
		params[name] = parts[i]
	}
	return params
}
