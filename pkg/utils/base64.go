package utils

import (
	"encoding/base64"
)

// DecodeCredentialRawID 直接使用 RawURLEncoding / URLEncoding 解碼
func DecodeCredentialRawID(base64URLStr string) ([]byte, error) {
	// 先嘗試使用不帶 padding (Raw) 的方式
	decoded, err := base64.RawURLEncoding.DecodeString(base64URLStr)
	if err == nil {
		return decoded, nil
	}
	// 如果失敗，再用帶 padding 的方式
	return base64.URLEncoding.DecodeString(base64URLStr)
}