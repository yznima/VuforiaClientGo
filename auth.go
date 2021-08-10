package vuforia

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
)

func prepare(secretKey, accessKey string, req *http.Request, body []byte) error {
	req.Header.Set("Date", time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	req.Header.Set("Content-Type", "application/json")

	signature, err := sign(secretKey, req, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("VWS %s:%s", accessKey, signature))

	return nil
}

// https://library.vuforia.com/articles/Training/Using-the-VWS-API.html
func sign(secretKey string, r *http.Request, body []byte) (string, error) {
	md5Hash := md5.New()
	if body != nil {
		_, err := md5Hash.Write(body)
		if err != nil {
			return "", err
		}
	}

	mac := hmac.New(sha1.New, []byte(secretKey))
	_, err := fmt.Fprintf(mac, "%s\n%x\n%s\n%s\n%s",
		r.Method,
		md5Hash.Sum(nil),
		r.Header.Get("Content-Type"),
		r.Header.Get("Date"),
		r.URL.Path,
	)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}
