package parser

import (
	"net/url"
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
	sourceURL   *url.URL
	publisher   string
}

func (a *Article) getMeta() {
	sources := []*meta{
		a.getJsonLD(),
		// getOgTags()
		// getTwitterTags()
		// getMetaTags()
	}

	a.getLang()
	a.getFavicon()
	a.getOriginURL()

	for _, s := range sources {
		metaEquals(&a.Author, s.author)
		metaEquals(&a.Title, s.title)
		metaEquals(&a.Description, s.description)
		metaEquals(&a.Favicon, s.favicon)
		metaEquals(&a.Poster, s.poster)
		metaEquals(&a.PublishedAt, s.publishedAt)
		metaEquals(&a.Publisher, s.publisher)
		metaEquals(&a.SourceURL, s.sourceURL)
	}

}

func (a *Article) getLang() {
	a.Lang = a.doc.Find("html").AttrOr("lang", "")
}

func (a *Article) getFavicon() {
	a.doc.Find("link[rel*='icon']").Each(func(i int, s *goquery.Selection) {
		href := s.AttrOr("href", "")

		fixed := fixLocalImg(href, a.baseURL)
		if href != "" && checkImg(fixed) {
			a.Favicon = fixed
			return
		}
	})
}

func (a *Article) getOriginURL() {
	a.doc.Find("link[rel='canonical']").Each(func(i int, s *goquery.Selection) {
		href := s.AttrOr("href", "")

		parsed, err := url.Parse(href)
		if rxURL.MatchString(href) && err == nil {
			a.SourceURL = parsed
		} else {
			a.SourceURL = a.baseURL
		}
	})
}

func metaEquals[T comparable](t *T, v T) {
	var zero T
	if *t == zero && v != zero {
		*t = v
	}
}
