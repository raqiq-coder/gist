package body

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var tagsToRemove = []string{"script", "noscript", "style", "iframe", "button", "br", "footer", "aside", "header", "nav", "details", "figcaption", "input", "textarea"}
var attrsToRemove = []string{"style", "template", "xmlns"}
var containerTags = []string{"div", "article", "section", "main", "p"}

type Body struct {
	Images []*Img
	HTML   *goquery.Document
	Text   string
	Len    int
}

type Img struct {
	Src string
	Alt string
}

func Extract(node *goquery.Document, baseURL *url.URL) (*Body, error) {
	clone := goquery.CloneDocument(node)
	body := clone.Find("body")
	if body.Contents().Length() == 0 {
		return nil, fmt.Errorf("failed to find document body")
	}

	preProcess(body)
	best := getBestCandidate(body)

	htmlContent, err := best.s.Html()
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	postProcessing(doc.Selection)
	fixImageSources(doc.Selection, baseURL)

	b := &Body{}
	if doc.Length() > 0 {
		b.Images = extractImages(doc.Selection)
		b.HTML = doc
		b.Text = doc.Text()
		b.Len = len(doc.Text())
	}

	return b, nil
}

func preProcess(node *goquery.Selection) {
	selector := strings.Join(tagsToRemove, ",")
	node.Find(selector).Remove()

	node.Find("*").Each(func(i int, s *goquery.Selection) {
		for _, attr := range attrsToRemove {
			s.RemoveAttr(attr)
		}
	})

	unwrapTags(node, "figure", "picture")
	removeEventListeners(node)
	removeEmptyTags(node)

	if html, err := node.Html(); err == nil {
		cleanHtml := rxHTMLComment.ReplaceAllString(html, "")
		node.SetHtml(cleanHtml)
	}
}

type candidate struct {
	s     *goquery.Selection
	score float64
}

func getBestCandidate(s *goquery.Selection) *candidate {
	var candidates []*candidate

	selector := strings.Join(containerTags, ",")
	s.Find(selector).Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		textLen := len(text)

		if textLen < 10 && s.Find("img").Length() == 0 {
			s.Remove()
		}

		if score := calcScore(s, textLen); score > 0 {
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
			if ratio > 0.85 {
				best = child
			}
		}
	}

	return best
}

func calcScore(s *goquery.Selection, textLen int) float64 {
	score := float64(textLen) / 100.0

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

func postProcessing(s *goquery.Selection) {
	s.Find("h4 + a[href*='.com']").Remove()
	s.Find("a").Each(func(i int, a *goquery.Selection) {
		href := a.AttrOr("href", "")
		if strings.HasPrefix(href, "#") {
			a.Remove()
		}

		// TODO: Нужно лучше продумать эту историю. А то сейчас удаляется и нормальный текст где есть спец символы
		linkText := a.Text()
		if hasSymbol(linkText) {
			a.Remove()
		}

		parsed, _ := url.Parse(href)
		if !parsed.IsAbs() {
			unwrapSelection(a)
		}
	})

	s.Find("h1, h2, h3, h4, h5, h6").FilterFunction(isSingleHeading).Remove()
	s.Find("hr").FilterFunction(isHrSeq).Remove()
	s.Find("svg").FilterFunction(isUIIcon).Remove()
	s.Find(strings.Join(tagsToRemove, ",")).Remove()

	s.Find("p").Each(func(i int, p *goquery.Selection) {
		text := p.Text()
		if rxDoantionText.MatchString(text) && len(text) < 100 {
			p.Remove()
		}
	})

	selector := strings.Join(containerTags, ",")
	s.Find(selector).Each(func(i int, s *goquery.Selection) {
		classID := getClassID(s)
		if matched := rxNegativeClasses.MatchString(classID); matched {
			s.Remove()
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

	s.Find("*").RemoveAttr("class").RemoveAttr("id")
	s.Find(`ul:has(a[href*="/page/"]), ol:has(a:contains("Next")), ol:has(a:contains("Previous"))`).Remove()

	removeDataAttrs(s)
	removeEmptyTags(s)
	unwrapTags(s, "main")
	unwrapNestedDivs(s)

	firstImg := s.Find("img").First()
	if isFirstImg(firstImg, 6) {
		firstImg.Remove()
	}

	formatHTML(s)
}

func fixImageSources(node *goquery.Selection, baseURL *url.URL) {
	attrsToRemove := []string{
		"srcset",
		"sizes",
		"data-srcset",
		"data-sizes",
		"loading",
		"decoding",
		"crossorigin",
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
			parsed, err := url.Parse(src)
			if err != nil {
				return
			}

			if realURL := parsed.Query().Get("url"); realURL != "" {
				decoded, err := url.QueryUnescape(realURL)
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
			fixImgSize(img)
		}
	})
}

func extractImages(node *goquery.Selection) []*Img {
	seen := map[string]any{}
	imgs := []*Img{}

	node.Find("img").Each(func(i int, s *goquery.Selection) {
		src, found := s.Attr("src")
		if !found || src == "" {
			return
		}

		if _, ok := seen[src]; !ok {
			seen[src] = struct{}{}

			imgs = append(imgs, &Img{
				Src: src,
				Alt: s.AttrOr("alt", ""),
			})
		}
	})

	return imgs
}
