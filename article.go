package parser

import (
	"fmt"
	nurl "net/url"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var tagsToRemove = []string{"script", "noscript", "style", "iframe", "button", "br", "footer", "aside", "header", "nav"}
var containerTags = []string{"div", "article", "section", "main", "p"}

type Article struct {
	Title       string
	Description string
	Author      string
	PublishedAt *time.Time
	Poster      string
	Favicon     string
	Lang        string
	SourceURL   string
	Publisher   string
	Content     *goquery.Document
	TextContent string
	Length      int
	Images      []*ImgMeta

	body *goquery.Selection
}

type ImgMeta struct {
	Src string
	Alt string
}

func (a *Article) extractArticleContent(baseURL *nurl.URL) {
	a.preProcessing()

	best := getBestCandidate(a.body)
	a.body = best.s

	a.postProcessing()
	a.fixImageSources(baseURL)

	doc := goquery.NewDocumentFromNode(a.body.Get(0))
	if doc.Length() > 0 {
		a.extractImages()
		a.Content = doc
		a.TextContent = best.s.Text()
		a.Length = len(best.s.Text())
	}
}

func (a *Article) preProcessing() {
	a.body.Find("*").RemoveAttr("style")

	selector := strings.Join(tagsToRemove, ",")
	a.body.Find(selector).Remove()

	unwrapTags(a.body, "figure", "picture")
	removeEmptyTags(a.body)
}

type candidate struct {
	s     *goquery.Selection
	score float64
}

func getBestCandidate(s *goquery.Selection) *candidate {
	var candidates []*candidate

	selector := strings.Join(containerTags, ",")
	s.Find(selector).Each(func(i int, s *goquery.Selection) {
		if score := calcScore(s); score > 0 {
			candidates = append(candidates, &candidate{s, score})
		}
	})

	if len(candidates) == 0 {
		return nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	best := candidates[0]
	for _, child := range candidates {
		if best.s.HasNodes(child.s.Get(0)) != nil {
			ratio := child.score / best.score
			if ratio > 0.8 {
				best = child
			}
		}
	}

	return best
}

func calcScore(s *goquery.Selection) float64 {
	text := strings.TrimSpace(s.Text())
	textLen := len(text)
	score := float64(textLen) / 100.0

	if textLen < 10 && s.Find("img").Length() == 0 {
		s.Remove()
	}

	linksTextLen := 0
	s.Find("a").Each(func(i int, a *goquery.Selection) {
		linksTextLen += len(strings.TrimSpace(a.Text()))
	})
	if textLen > 0 {
		linkDensity := float64(linksTextLen) / float64(textLen)
		score *= (1 - linkDensity)
	}

	if pCount := s.Find("p").Length(); pCount > 0 {
		score += float64(pCount) * 1.5
	}

	depth := 0
	s.Parents().Each(func(i int, p *goquery.Selection) { depth++ })
	if depth > 10 {
		score *= 0.9
	}

	tag := s.Get(0).Data
	if tag == "article" || tag == "main" {
		score += 20.0
	}

	class, _ := s.Attr("class")
	id, _ := s.Attr("id")
	classID := strings.ToLower(class + " " + id)

	if matched := rxPositiveClasses.MatchString(classID); matched {
		score += 20.0
	}

	if matched := rxNegativeClasses.MatchString(classID); matched {
		score -= 20.0
		s.Remove()
	}

	return score
}

func (a *Article) postProcessing() {
	s := a.body

	s.Find("h1, h2, h3, h4, h5, h6").Each(func(i int, h *goquery.Selection) {
		after := h.NextAll()
		hasContent := false

		after.Each(func(j int, el *goquery.Selection) {
			if len(strings.TrimSpace(el.Text())) > 20 || el.Is("img, ul, ol, table, blockquote") {
				hasContent = true
				return
			}
		})

		parent := h.Parent()
		ignoreParents := []string{
			"body", "article", "main", "section",
		}

		isRootParent := slices.Contains(ignoreParents, parent.Get(0).Data)
		if parent.Length() > 0 && !isRootParent {
			parent.NextAll().Each(func(j int, el *goquery.Selection) {
				if len(strings.TrimSpace(el.Text())) > 20 || el.Is("img, ul, ol, table, blockquote") {
					hasContent = true
					return
				}
			})
		}

		if h.Prev().Length() == 0 && h.Parent().Prev().Length() == 0 {
			h.Remove()
		}

		if !hasContent {
			h.Remove()
		}
	})

	s.Find("*").RemoveAttr("class").RemoveAttr("id")
	s.Find(`ul:has(a[href*="/page/"]), ol:has(a:contains("Next")), ol:has(a:contains("Previous"))`).Remove()
	s.Find(`[class*="pagination"], [class*="pager"], 
		[class*="post-navigation"], [class*="page-nav"],
		[class*="related"], [class*="comment"],
		[class*="share"], [class*="social"],
		[class*="banner"], [class*="nav"]`).Remove()

	removeEventListeners(s)
	removeDataAttrs(s)
	removeEmptyTags(s)

	s.Find("hr").Each(func(i int, hr *goquery.Selection) {
		next := hr.Next()
		prev := hr.Prev()

		if next.Length() == 0 || prev.Length() == 0 {
			hr.Remove()
		}

		if next.Is("hr") || prev.Is("hr") {
			hr.Remove()
		}
	})

	html, _ := s.Html()
	cleanHtml := removeSpace(html)
	cleanHtml = rxHTMLComment.ReplaceAllString(cleanHtml, "")

	s.SetHtml(cleanHtml)
}

func (a *Article) fixImageSources(baseURL *nurl.URL) {
	attrsToRemove := []string{
		"srcset",
		"sizes",
		"data-srcset",
		"data-sizes",
		"loading",
		"decoding",
		"crossorigin",
		"width",
		"height",
	}

	a.body.Find("img").Each(func(i int, img *goquery.Selection) {
		for _, attr := range attrsToRemove {
			img.RemoveAttr(attr)
		}

		src, exists := img.Attr("src")
		if !exists || src == "" {
			return
		}

		///_next/image?url=
		if strings.HasPrefix(src, "/_next/image") {
			parsed, err := nurl.Parse(src)
			if err == nil {
				if realURL := parsed.Query().Get("url"); realURL != "" {
					decoded, err := nurl.QueryUnescape(realURL)
					if err == nil {
						src = decoded
					} else {
						src = realURL
					}
				}
			}
		}

		if !strings.HasPrefix(src, "http") && strings.HasPrefix(src, "/") && !strings.HasPrefix(src, "//") {
			src = fmt.Sprintf("%s://%s/%s", baseURL.Scheme, baseURL.Host, src)
		}

		img.SetAttr("src", src)
		img.SetAttr("width", "100%")
		img.SetAttr("height", "auto")
	})
}

func (a *Article) extractImages() {
	seen := map[string]any{}

	a.body.Find("img").Each(func(i int, s *goquery.Selection) {
		src, found := s.Attr("src")
		if !found || src == "" {
			return
		}

		if _, ok := seen[src]; !ok {
			seen[src] = struct{}{}

			a.Images = append(a.Images, &ImgMeta{
				Src: src,
				Alt: s.AttrOr("alt", ""),
			})
		}
	})
}

func (a *Article) PrintMeta() {
	fmt.Println("OriginURL: ", a.SourceURL)
	fmt.Println("Author: ", a.Author)
	fmt.Println("Title: ", a.Title)
	fmt.Println("Description: ", a.Description)
	fmt.Println("Poster: ", a.Poster)
	fmt.Println("PublishedAt: ", a.PublishedAt)
	fmt.Println("Publisher: ", a.Publisher)
	fmt.Println("Favicon: ", a.Favicon)
	fmt.Println("Languange: ", a.Lang)
}
