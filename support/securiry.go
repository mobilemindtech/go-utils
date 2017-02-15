package support

import (
  "crypto/sha1"
  "crypto/sha256"
  "encoding/base64"  
)

func TextToSha1(text string) string{
  bv := []byte(text) 
  hasher := sha1.New()
  hasher.Write(bv)
  sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))	
  return sha
}

func IsSameHash(hash string , text string) bool {
	newHash := TextToSha1(text)
	return newHash == hash
}

func TextToSha256(text string) string{
  bv := []byte(text) 
  hasher := sha256.New()
  hasher.Write(bv)
  sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))	
  return sha
}

func IsSameHashSha256(hash string , text string) bool {
	newHash := TextToSha256(text)
	return newHash == hash
}