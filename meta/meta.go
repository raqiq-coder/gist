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

func Extract(node *goquery.Selection, baseURL *nurl.URL) *Meta {
	m := &Meta{}

	sources := []*Meta{
		jsonLD(node),
		ogTags(node),
		twitterTags(node),
		metaHTML(node),
	}

	for _, s := range sources {
		metaEquals(&m.Author, s.Author)
		metaEquals(&m.Title, s.Title)
		metaEquals(&m.Description, s.Description)
		metaEquals(&m.Favicon, s.Favicon)
		metaEquals(&m.Poster, s.Poster)
		metaEquals(&m.PublishedAt, s.PublishedAt)
		metaEquals(&m.Publisher, s.Publisher)
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

func metaEquals[T comparable](t *T, v T) {
	var zero T
	if *t == zero && v != zero {
		*t = v
	}
}
