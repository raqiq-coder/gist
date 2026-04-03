package meta

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

type ogTags struct{}

var _ metasource = &ogTags{}

func (*ogTags) extract(node *goquery.Selection) *Meta {
	m := &Meta{}

	for k, v := range ogSelectors {
		sel := fmt.Sprintf("meta[property='og:%s']", v)

		node.Find(sel).Each(func(i int, s *goquery.Selection) {
			val := s.AttrOr("content", "")
			m.setMetaField(k, val)
		})
	}

	authorVal := node.Find("meta[property='article:author']").AttrOr("content", "")
	m.setMetaField(author, authorVal)

	publishedAtVal := node.Find("meta[property='article:published_time']").AttrOr("content", "")
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

type twitterTags struct{}

var _ metasource = &twitterTags{}

func (*twitterTags) extract(node *goquery.Selection) *Meta {
	m := &Meta{}

	for k, v := range twitterSelectors {
		sel := fmt.Sprintf("meta[name='twitter:%s']", v)

		node.Find(sel).Each(func(i int, s *goquery.Selection) {
			val := s.AttrOr("content", "")
			m.setMetaField(k, val)
		})
	}

	return m
}

type htmlTags struct{}

var _ metasource = &htmlTags{}

func (*htmlTags) extract(node *goquery.Selection) *Meta {
	m := &Meta{}

	authorVal := node.Find("meta[name='author']").AttrOr("content", "")
	m.setMetaField(author, authorVal)

	descVal := node.Find("meta[name='description']").AttrOr("content", "")
	m.setMetaField(description, descVal)

	titleVal := node.Find("html title").Text()
	m.setMetaField(title, titleVal)

	return m
}

func (m *Meta) setMetaField(tagType string, val string) {
	switch tagType {
	case title:
		m.Title = val
	case description:
		m.Description = val
	case poster:
		m.Poster = val
	case publisher:
		m.Publisher = val
	case site:
		m.Publisher = val
	case author:
		m.Author = val
	case creator:
		m.Author = val
	case sourceURL:
		parsed, err := nurl.Parse(val)
		if err == nil {
			m.SourceURL = parsed
		}
	case publishedAt:
		if val != "" {
			parsed, err := time.Parse(time.RFC3339, val)
			if err == nil {
				m.PublishedAt = &parsed
			}
		}
	}

}
