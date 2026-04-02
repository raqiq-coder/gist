package parser

import (
	nurl "net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type meta struct {
	title       string
	description string
	author      string
	publishedAt *time.Time
	poster      string
	favicon     string
	sourceURL   *nurl.URL
	publisher   string
	lang        string
}

func (p *Parser) extractMeta() *meta {
	m := &meta{}

	sources := []*meta{
		jsonLD(p.doc),
		ogTags(p.doc),
		twitterTags(p.doc),
		metaHTML(p.doc),
	}

	for _, s := range sources {
		metaEquals(&m.author, s.author)
		metaEquals(&m.title, s.title)
		metaEquals(&m.description, s.description)
		metaEquals(&m.favicon, s.favicon)
		metaEquals(&m.poster, s.poster)
		metaEquals(&m.publishedAt, s.publishedAt)
		metaEquals(&m.publisher, s.publisher)
	}

	m.lang = p.lang()
	m.favicon = p.favicon()
	m.sourceURL = p.originURL()

	return m
}

func (p *Parser) lang() string {
	return p.doc.Find("html").AttrOr("lang", "")
}

func (p *Parser) favicon() string {
	icon := ""
	p.doc.Find("link[rel*='icon']").Each(func(i int, s *goquery.Selection) {
		href := s.AttrOr("href", "")

		fixed := fixLocalImg(href, p.baseURL)
		if href != "" && checkImg(fixed) {
			icon = fixed
			return
		}
	})

	return icon
}

func (p *Parser) originURL() *nurl.URL {
	var url *nurl.URL
	p.doc.Find("link[rel='canonical']").Each(func(i int, s *goquery.Selection) {
		href := s.AttrOr("href", "")

		parsed, err := nurl.Parse(href)
		if rxURL.MatchString(href) && err == nil {
			url = parsed
		} else {
			url = p.baseURL
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
