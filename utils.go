package parser

import (
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func unwrapTags(root *goquery.Selection, tags ...string) {
	selector := strings.Join(tags, ",")
	root.Find(selector).Each(func(i int, s *goquery.Selection) {
		unwrapSelection(s)
	})
}

func unwrapSelection(s *goquery.Selection) {
	if s.Length() == 0 {
		return
	}

	contents := s.Contents()
	if contents.Length() > 0 {
		s.ReplaceWithNodes(contents.Nodes...)
	} else {
		s.Remove()
	}
}

func removeEmptyTags(root *goquery.Selection) {
	changed := true
	for changed {
		changed = false

		root.Find("*").Each(func(i int, s *goquery.Selection) {
			tag := s.Get(0).Data
			if tag == "br" || tag == "hr" || tag == "img" {
				return
			}

			text := strings.TrimSpace(s.Text())
			hasMedia := s.Find("img, video, audio, iframe").Length() > 0

			if text == "" && !hasMedia {
				s.Remove()
				changed = true
			}
		})
	}
}

func removeDataAttrs(root *goquery.Selection) {
	root.Find("[data-]").Each(func(i int, s *goquery.Selection) {
		node := s.Get(0)
		if node == nil {
			return
		}

		for _, attr := range node.Attr {
			if strings.HasPrefix(attr.Key, "data-") {
				s.RemoveAttr(attr.Key)
			}
		}
	})
}

func removeEventListeners(root *goquery.Selection) {
	html, err := root.Html()
	if err != nil || html == "" {
		return
	}

	cleanHtml := rxEventAttrs.ReplaceAllString(html, "")

	root.SetHtml(cleanHtml)
}

func removeSpace(html string) string {
	html = rxSpaceBetweenTags.ReplaceAllString(html, "><")
	html = rxMultipleSpaces.ReplaceAllString(html, " ")

	return strings.TrimSpace(html)
}

func checkImg(img string) bool {
	parsed, err := url.Parse(img)
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

func fixLocalImg(relativePath string, base *url.URL) string {
	fullURL, err := base.Parse(relativePath)
	if err != nil {
		return ""
	}

	fullURL.Path = filepath.Clean(fullURL.Path)

	return fullURL.String()
}

func hasSymbol(str string) bool {
	return strings.ContainsAny(str, "!#$%^&*()_+{}[]")
}
