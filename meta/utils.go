package meta

import (
	"net/http"
	nurl "net/url"
	"path/filepath"
	"strings"
	"time"
)

func checkImg(img string) bool {
	parsed, err := nurl.Parse(img)
	if err != nil {
		return false
	}

	req, err := http.NewRequest(http.MethodHead, parsed.String(), nil)
	if err != nil {
		return false
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ImgChecker/1.0)")

	client := http.Client{
		Timeout: 3 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	res, err := client.Do(req)
	if err != nil {
		return false
	}
	defer res.Body.Close()

	contentType := res.Header.Get("Content-Type")
	isSuccess := res.StatusCode == 200 ||
		res.StatusCode == 304 ||
		(res.StatusCode >= 301 && res.StatusCode <= 308)

	return isSuccess && strings.HasPrefix(contentType, "image/")
}

func fixLocalImg(relativePath string, base *nurl.URL) string {
	fullURL, err := base.Parse(relativePath)
	if err != nil {
		return ""
	}

	fullURL.Path = filepath.Clean(fullURL.Path)

	return fullURL.String()
}
