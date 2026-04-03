package meta

import (
	nurl "net/url"
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
	SourceURL   *nurl.URL
	Publisher   string
	Lang        string
}

type metasource interface {
	extract(node *goquery.Selection) *Meta
}

func Extract(node *goquery.Selection, baseURL *nurl.URL) *Meta {
	msources := []metasource{
		&jsonLD{},
		&ogTags{},
		&twitterTags{},
		&htmlTags{},
	}

	m := &Meta{}
	for _, ms := range msources {
		res := ms.extract(node)
		assignMeta(m, res)
	}

	m.Lang = lang(node)
	m.Favicon = favicon(node, baseURL)
	m.SourceURL = originURL(node, baseURL)

	return m
}

func lang(node *goquery.Selection) string {
	return node.Find("html").AttrOr("lang", "")
}

func favicon(node *goquery.Selection, baseURL *nurl.URL) string {
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

func originURL(node *goquery.Selection, baseURL *nurl.URL) *nurl.URL {
	var url *nurl.URL
	node.Find("link[rel='canonical']").Each(func(i int, s *goquery.Selection) {
		href := s.AttrOr("href", "")

		parsed, err := nurl.Parse(href)
		if rxURL.MatchString(href) && err == nil {
			url = parsed
		} else {
			url = baseURL
		}
	})

	return url
}

func assignMeta(t, s *Meta) {
	metaEquals(&t.Author, s.Author)
	metaEquals(&t.Title, s.Title)
	metaEquals(&t.Description, s.Description)
	metaEquals(&t.Favicon, s.Favicon)
	metaEquals(&t.Poster, s.Poster)
	metaEquals(&t.PublishedAt, s.PublishedAt)
	metaEquals(&t.Publisher, s.Publisher)
}

func metaEquals[T comparable](t *T, s T) {
	var zero T
	if *t == zero && s != zero {
		*t = s
	}
}
