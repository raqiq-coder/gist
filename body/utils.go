package body

import (
	nurl "net/url"
	"path/filepath"
	"strings"

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

func hasSymbol(str string) bool {
	return strings.ContainsAny(str, "!#$%^&*()_+{}[]")
}

func getClassID(s *goquery.Selection) string {
	class := s.AttrOr("class", "")
	id := s.AttrOr("id", "")

	return strings.ToLower(class + " " + id)
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

func fixLocalImg(relativePath string, base *nurl.URL) string {
	fullURL, err := base.Parse(relativePath)
	if err != nil {
		return ""
	}

	fullURL.Path = filepath.Clean(fullURL.Path)
	return fullURL.String()
}
