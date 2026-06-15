package appium

import "encoding/base64"

// base64Decode, WebDriver'ın döndüğü base64 string'i decode eder
func base64Decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}
