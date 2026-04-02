package meta

import (
	"encoding/json"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func jsonLD(node *goquery.Selection) *Meta {
	meta := &Meta{}

	node.Find(`script[type="application/ld+json"]`).Each(func(i int, s *goquery.Selection) {
		content := rxCDATA.ReplaceAllString(s.Text(), "")

		var parsed map[string]any
		err := json.Unmarshal([]byte(content), &parsed)
		if err != nil {
			return
		}

		ldContext, isString := parsed["@context"].(string)
		if !isString || !rxSchemaOrg.MatchString(ldContext) {
			return
		}

		if _, typeExist := parsed["@type"]; !typeExist {
			graphList, isArray := parsed["@graph"].([]any)
			if !isArray {
				return
			}

			for _, graph := range graphList {
				objGraph, isObj := graph.(map[string]any)
				if !isObj {
					continue
				}

				strType, isString := objGraph["@type"].(string)
				if isString && rxJsonLdArticleTypes.MatchString(strType) {
					parsed = objGraph
					break
				}
			}
		}

		ldType, isString := parsed["@type"].(string)
		if !isString || !rxJsonLdArticleTypes.MatchString(ldType) {
			return
		}

		meta.Title = getLdString(parsed["headline"])
		meta.Description = getLdString(parsed["description"])
		meta.Author = getLdPerson(parsed["author"])
		meta.Publisher = getLdPerson(parsed["publisher"])
		meta.Poster = getLdImage(parsed["image"])
		meta.PublishedAt = getLdPublishedAt(parsed["datePublished"])
	})

	return meta
}

func getLdString(source any) string {
	ldString, isString := source.(string)
	if !isString {
		return ""
	}

	return ldString

}

func getLdPerson(source any) string {
	switch p := source.(type) {
	case map[string]any:
		if name, isString := p["name"].(string); isString {
			return name
		}

	case string:
		return p
	}

	return ""
}

func getLdImage(source any) string {
	switch data := source.(type) {
	case string:
		if checkImg(data) {
			return data
		}

	case []any:
		for _, img := range data {
			imgUrl, isString := img.(string)
			if isString && checkImg(imgUrl) {
				return imgUrl
			}
		}

	case map[string]any:
		url, isString := data["url"].(string)
		if isString && checkImg(url) {
			return url
		}
	}

	return ""
}

func getLdPublishedAt(source any) *time.Time {
	switch date := source.(type) {
	case string:
		if t, err := time.Parse(time.RFC3339, date); err == nil {
			return &t
		}

		if t, err := time.Parse(time.RFC3339Nano, date); err == nil {
			return &t
		}

		if t, err := time.Parse("2006-01-02", date); err == nil {
			return &t
		}

		return nil

	case float64:
		t := time.Unix(int64(date), 0)
		return &t

	default:
		return nil
	}
}
