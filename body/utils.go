package body

import (
	nurl "net/url"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func unwrapTags(root *goquery.Selection, tags ...string) {
	selector := strings.Join(tags, ",")
	root.Find(selector).Each(func(i int, s *goquery.Selection) {
		unwrapSelection(s)
	})
}

func unwrapSelection(s *goquery.Selection) {
	if s.Length() == 0 {
		return
	}

	contents := s.Contents()
	if contents.Length() > 0 {
		s.ReplaceWithNodes(contents.Nodes...)
	} else {
		s.Remove()
	}
}

func unwrapNestedDivs(root *goquery.Selection) {
	divs := root.Find("div")
	for i := divs.Length() - 1; i >= 0; i-- {
		div := divs.Eq(i)
		if div.Children().Length() == 1 {
			unwrapSelection(div)
		}
	}
}

func removeEmptyTags(root *goquery.Selection) {
	changed := true
	for changed {
		changed = false

		root.Find("*").Each(func(i int, s *goquery.Selection) {
			tag := s.Get(0).Data
			if tag == "br" || tag == "hr" || tag == "img" {
				return
			}

			text := strings.TrimSpace(s.Text())
			hasMedia := s.Find("img, video, audio, iframe").Length() > 0

			if text == "" && !hasMedia {
				s.Remove()
				changed = true
			}
		})
	}
}

func hasSymbol(str string) bool {
	return strings.ContainsAny(str, "!#$^&*_+{}[]")
}

func getClassID(s *goquery.Selection) string {
	class := s.AttrOr("class", "")
	id := s.AttrOr("id", "")

	return strings.ToLower(class + " " + id)
}

func removeDataAttrs(root *goquery.Selection) {
	root.Find("*").Each(func(i int, s *goquery.Selection) {
		attrs := s.Get(0).Attr
		for _, attr := range attrs {
			if strings.HasPrefix(attr.Key, "data-") {
				s.RemoveAttr(attr.Key)
			}
		}
	})
}

func removeEventListeners(root *goquery.Selection) {
	html, err := root.Html()
	if err != nil || html == "" {
		return
	}

	cleanHtml := rxEventAttrs.ReplaceAllString(html, "")
	root.SetHtml(cleanHtml)
}

func fixLocalImg(relativePath string, base *nurl.URL) string {
	fullURL, err := base.Parse(relativePath)
	if err != nil {
		return ""
	}

	fullURL.Path = filepath.Clean(fullURL.Path)
	return fullURL.String()
}

func fixImgSize(img *goquery.Selection) {
	w, err := strconv.Atoi(img.AttrOr("width", ""))
	if err != nil {
		w = 100
	}

	h, err := strconv.Atoi(img.AttrOr("height", ""))
	if err != nil {
		h = 100
	}

	if w < 100 || h < 100 {
		return
	}

	img.SetAttr("width", "100%")
	img.SetAttr("height", "auto")
}

func isUIIcon(_ int, s *goquery.Selection) bool {
	width := s.AttrOr("width", "")
	height := s.AttrOr("height", "")
	if (strings.Contains(width, "px") && len(width) < 4) ||
		(strings.Contains(height, "px") && len(height) < 4) {
		return true
	}

	viewBox, _ := s.Attr("viewBox")
	if strings.HasPrefix(viewBox, "0 0 2") { // 24x24, 20x20
		return true
	}

	if s.Find("title").Length() != 0 {
		return true
	}

	return false
}

func isHrSeq(_ int, s *goquery.Selection) bool {
	next := s.Next()
	prev := s.Prev()

	if next.Length() == 0 || prev.Length() == 0 {
		return true
	}

	if next.Is("hr") || prev.Is("hr") {
		return true
	}

	return false
}

func isSingleHeading(_ int, s *goquery.Selection) bool {
	after := s.NextAll()
	hasContent := false

	after.Each(func(j int, el *goquery.Selection) {
		if len(strings.TrimSpace(el.Text())) > 20 || el.Is("img, ul, ol, table, blockquote") {
			hasContent = true
			return
		}
	})

	parent := s.Parent()
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

	if s.Prev().Length() == 0 && s.Parent().Prev().Length() == 0 {
		return true
	}

	if !hasContent {
		return true
	}

	return false
}

func isFirstImg(s *goquery.Selection, depth int) bool {
	current := s

	hasNeighbour := s.Prev().Length() != 0
	if hasNeighbour {
		return false
	}

	for i := 0; i < depth; i++ {
		parent := current.Parent()

		if parent.Length() == 0 {
			break
		}

		if parent.Prev().Length() != 0 {
			return false
		}

		current = parent
	}

	return true
}

func formatHTML(s *goquery.Selection) {
	s.Contents().Each(func(i int, child *goquery.Selection) {
		formatHTML(child)
	})

	removeEmptyTextNodes(s)
}

func removeEmptyTextNodes(s *goquery.Selection) {
	var toRemove []*html.Node

	s.Contents().Each(func(i int, child *goquery.Selection) {
		node := child.Get(0)
		if node == nil {
			return
		}

		if node.Type == html.TextNode {
			text := strings.TrimSpace(node.Data)

			if node.Parent != nil {
				parentTag := node.Parent.Data
				if parentTag == "code" || parentTag == "pre" {
					return
				}
			}

			if text == "" {
				toRemove = append(toRemove, node)
			}
		}
	})

	for _, node := range toRemove {
		node.Parent.RemoveChild(node)
	}
}
