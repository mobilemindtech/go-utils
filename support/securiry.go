package support

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	beego "github.com/beego/beego/v2/server/web"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/mobilemindtec/go-io/result"
	"math/rand"
	"time"
)

func TextToSha1(text string) string {
	bv := []byte(text)
	hasher := sha1.New()
	hasher.Write(bv)
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}

func TextToSha1Hex(text string) string {
	bv := []byte(text)
	hasher := sha1.New()
	hasher.Write(bv)
	sha := hex.EncodeToString(hasher.Sum(nil))
	return sha
}

func IsSameHash(hash string, text string) bool {
	newHash := TextToSha1(text)
	return newHash == hash
}

func IsSameHashHex(hash string, text string) bool {
	newHash := TextToSha1Hex(text)
	return newHash == hash
}

func TextToSha256(text string) string {
	bv := []byte(text)
	hasher := sha256.New()
	hasher.Write(bv)
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}

func TextToSha256Hex(text string) string {
	bv := []byte(text)
	hasher := sha256.New()
	hasher.Write(bv)
	sha := hex.EncodeToString(hasher.Sum(nil))
	return sha
}

func IsSameHashSha256(hash string, text string) bool {
	newHash := TextToSha256(text)
	return newHash == hash
}

func IsSameHashSha256Hex(hash string, text string) bool {
	newHash := TextToSha256Hex(text)
	return newHash == hash
}

func NewBearerToken(data jwt.MapClaims, secret string) *result.Result[string] {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	hmacSampleSecret := []byte(secret)
	return result.Try(func() (string, error) {
		return token.SignedString(hmacSampleSecret)
	})
}

func GenereteApiToken(id int64, uuid string, password string, expirationDate time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":         id,
		"user_uuid":       uuid,
		"user_password":   password,
		"expiration_date": expirationDate.Unix(),
	})

	secret, _ := beego.AppConfig.String("jwt_token_secret")
	hmacSampleSecret := []byte(secret)
	return token.SignedString(hmacSampleSecret)
}

func GenereteToken(password string, expirationDate time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"password":        password,
		"expiration_date": expirationDate.Unix(),
	})

	secret, _ := beego.AppConfig.String("jwt_token_secret")
	hmacSampleSecret := []byte(secret)
	return token.SignedString(hmacSampleSecret)
}

func GenerateCode(min int, max int) int {
	return min + rand.Intn(max-min)
}
