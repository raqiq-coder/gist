package meta

import (
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Meta struct {
	Title       string
	Description string
	Author      string
	PublishedAt *time.Time
	Poster      string
	Favicon     string
	SourceURL   *url.URL
	Publisher   string
	Lang        string
}

type metasource interface {
	extract(node *goquery.Selection) *Meta
}

func Extract(node *goquery.Selection, baseURL *url.URL) *Meta {
	msources := []metasource{
		&ogTags{},
		&jsonLD{},
		&twitterTags{},
		&htmlTags{},
	}

	m := &Meta{}
	m.Lang = lang(node)
	m.Favicon = favicon(node, baseURL)
	m.SourceURL = originURL(node, baseURL)

	for _, ms := range msources {
		if isFilled(*m) {
			break
		}

		res := ms.extract(node)
		assignMeta(m, res)
	}

	return m
}

func lang(node *goquery.Selection) string {
	return node.Find("html").AttrOr("lang", "")
}

func favicon(node *goquery.Selection, baseURL *url.URL) string {
	icon := ""
	node.Find("link[rel*='icon']").Each(func(i int, s *goquery.Selection) {
		href := s.AttrOr("href", "")

		fixed := fixLocalImg(href, baseURL)
		if href != "" && checkImg(fixed) {
			icon = fixed
			return
		}
	})

	return icon
}

func originURL(node *goquery.Selection, baseURL *url.URL) *url.URL {
	var url *url.URL
	node.Find("link[rel='canonical']").Each(func(i int, s *goquery.Selection) {
		href := s.AttrOr("href", "")

		parsed, err := url.Parse(href)
		if rxURL.MatchString(href) && err == nil {
			url = parsed
		} else {
			url = baseURL
		}
	})

	return url
}

func assignMeta(t, s *Meta) {
	equateNonEmpty(&t.Author, s.Author)
	equateNonEmpty(&t.Title, s.Title)
	equateNonEmpty(&t.Description, s.Description)
	equateNonEmpty(&t.Favicon, s.Favicon)
	equateNonEmpty(&t.Poster, s.Poster)
	equateNonEmpty(&t.PublishedAt, s.PublishedAt)
	equateNonEmpty(&t.Publisher, s.Publisher)
}
