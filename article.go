package parser

import (
	"fmt"
	nurl "net/url"
	"slices"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var tagsToRemove = []string{"script", "noscript", "style", "iframe", "button", "br", "footer", "aside", "header", "nav", "details", "figcaption", "input", "textarea"}
var containerTags = []string{"div", "article", "section", "main", "p"}

type content struct {
	images []*ImgMeta
	html   *goquery.Document
	text   string
	len    int
}

type ImgMeta struct {
	Src string
	Alt string
}

func (p *Parser) extractContent() (*content, error) {
	body := p.doc.Find("body")
	if body.Contents().Length() == 0 {
		return nil, fmt.Errorf("failed to find document body")
	}

	p.doc = body

	p.preProcessing()

	best := getBestCandidate(p.doc)
	p.doc = best.s

	p.postProcessing()
	fixImageSources(p.doc, p.baseURL)

	c := &content{}

	doc := goquery.NewDocumentFromNode(p.doc.Get(0))
	if doc.Length() > 0 {
		c.images = p.extractImages()
		c.html = doc
		c.text = best.s.Text()
		c.len = len(best.s.Text())
	}

	return c, nil
}

func (p *Parser) preProcessing() {
	p.doc.Find("*").RemoveAttr("style")

	selector := strings.Join(tagsToRemove, ",")
	p.doc.Find(selector).Remove()

	unwrapTags(p.doc, "figure", "picture")
	removeEmptyTags(p.doc)
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

		classID := getClassID(s)
		if matched := rxNegativeClasses.MatchString(classID); matched {
			s.Remove()
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
			if ratio > 0.85 {
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
		score *= (1 - linkDensity) + 100
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

	classID := getClassID(s)

	if matched := rxPositiveClasses.MatchString(classID); matched {
		score += 20.0
	}

	if matched := rxNegativeClasses.MatchString(classID); matched {
		score -= 20.0
	}

	return score
}

func (p *Parser) postProcessing() {
	s := p.doc

	s.Find("h4 + a[href*='.com']").Remove()
	s.Find("a").Each(func(i int, a *goquery.Selection) {
		href := a.AttrOr("href", "")
		if strings.HasPrefix(href, "#") {
			a.Remove()
		}

		linkText := a.Text()
		if hasSymbol(linkText) {
			a.Remove()
		}

		parsed, _ := nurl.Parse(href)
		if !parsed.IsAbs() {
			unwrapSelection(a)
		}
	})

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

	s.Find("a, div").Each(func(i int, a *goquery.Selection) {
		contents := a.Contents()
		if contents.Length() == 1 {
			first := contents.First()
			if first.Is("img") {
				a.ReplaceWithSelection(first.Clone())
			}
		}
	})

	firstImg := s.Find("img").First()
	unwrapTags(firstImg, "div")
	if firstImg.Prev().Length() == 0 && firstImg.Parent().Prev().Length() == 0 {
		firstImg.Remove()
	}

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

	html, _ := s.Html()
	cleanHtml := removeSpace(html)
	cleanHtml = rxHTMLComment.ReplaceAllString(cleanHtml, "")

	s.SetHtml(cleanHtml)
}

func fixImageSources(node *goquery.Selection, baseURL *nurl.URL) {
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

	node.Find("img").Each(func(i int, img *goquery.Selection) {
		for _, attr := range attrsToRemove {
			img.RemoveAttr(attr)
		}

		src, exists := img.Attr("src")
		if !exists || src == "" {
			return
		}

		var fullSrc string

		///_next/image?url=
		if strings.HasPrefix(src, "/_next/image") {
			parsed, err := nurl.Parse(src)
			if err != nil {
				return
			}

			if realURL := parsed.Query().Get("url"); realURL != "" {
				decoded, err := nurl.QueryUnescape(realURL)
				if err == nil {
					fullSrc = decoded
				} else {
					fullSrc = realURL
				}
			}
		} else {
			fullSrc = fixLocalImg(src, baseURL)
		}

		if fullSrc != "" {
			img.SetAttr("src", fullSrc)
			img.SetAttr("width", "100%")
			img.SetAttr("height", "auto")
		}
	})
}

func (p *Parser) extractImages() []*ImgMeta {
	seen := map[string]any{}
	imgs := []*ImgMeta{}

	p.doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, found := s.Attr("src")
		if !found || src == "" {
			return
		}

		if _, ok := seen[src]; !ok {
			seen[src] = struct{}{}

			imgs = append(imgs, &ImgMeta{
				Src: src,
				Alt: s.AttrOr("alt", ""),
			})
		}
	})

	return imgs
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
