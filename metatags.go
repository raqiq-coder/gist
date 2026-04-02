package parser

import (
	"fmt"
	"time"

	"github.com/PuerkitoBio/goquery"

	nurl "net/url"
)

const (
	// og
	title       = "title"
	description = "description"
	poster      = "poster"
	sourceURL   = "sourceURL"
	publisher   = "publisher"
	publishedAt = "publishedAt"
	author      = "author"

	// twitter
	site    = "site"
	creator = "creator"
)

var ogSelectors = map[string]string{
	title:       "title",
	description: "description",
	poster:      "image",
	sourceURL:   "url",
	publisher:   "site_name",
}

func (a *Article) getOgTags() *meta {
	m := &meta{}

	for k, v := range ogSelectors {
		sel := fmt.Sprintf("meta[property='og:%s']", v)

		a.doc.Find(sel).Each(func(i int, s *goquery.Selection) {
			val := s.AttrOr("content", "")
			m.setMetaField(k, val)
		})
	}

	authorVal := a.doc.Find("meta[property='article:author']").AttrOr("content", "")
	m.setMetaField(author, authorVal)

	publishedAtVal := a.doc.Find("meta[property='article:published_time']").AttrOr("content", "")
	m.setMetaField(publishedAt, publishedAtVal)

	return m
}

var twitterSelectors = map[string]string{
	title:       "title",
	description: "description",
	poster:      "image",
	site:        "site",
	creator:     "creator",
}

func (a *Article) getTwitterTags() *meta {
	m := &meta{}

	for k, v := range twitterSelectors {
		sel := fmt.Sprintf("meta[name='twitter:%s']", v)

		a.doc.Find(sel).Each(func(i int, s *goquery.Selection) {
			val := s.AttrOr("content", "")
			m.setMetaField(k, val)
		})
	}

	return m
}

func (a *Article) getMetaHTML() *meta {
	m := &meta{}

	authorVal := a.doc.Find("meta[name='author']").AttrOr("content", "")
	m.setMetaField(author, authorVal)

	descVal := a.doc.Find("meta[name='description']").AttrOr("content", "")
	m.setMetaField(description, descVal)

	titleVal := a.doc.Find("html title").Text()
	m.setMetaField(title, titleVal)

	return m
}

func (m *meta) setMetaField(tagType string, val string) {
	switch tagType {
	case title:
		m.title = val
	case description:
		m.description = val
	case poster:
		m.poster = val
	case publisher:
		m.publisher = val
	case site:
		m.publisher = val
	case author:
		m.author = val
	case creator:
		m.author = val
	case sourceURL:
		parsed, err := nurl.Parse(val)
		if err == nil {
			m.sourceURL = parsed
		}
	case publishedAt:
		if val != "" {
			parsed, err := time.Parse(time.RFC3339, val)
			if err == nil {
				m.publishedAt = &parsed
			}
		}
	}

}
