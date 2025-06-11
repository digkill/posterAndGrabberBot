package helpers

import (
	"crypto/md5"
	"encoding/hex"
)

func UrlKey(url string) []byte {
	h := md5.Sum([]byte(url))
	return []byte("processed_" + hex.EncodeToString(h[:]))
}
