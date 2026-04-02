package gist

import (
	"fmt"
	"net/http"
	nurl "net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/raqiq-coder/gist/body"
	"github.com/raqiq-coder/gist/meta"
)

type Parser struct {
	cfg *ParserCfg
}

type ParserCfg struct {
	MaxItemsLen int

	Timeout   time.Duration
	UserAgent string
}

func NewParser(cfg *ParserCfg) *Parser {
	if cfg == nil {
		cfg = &ParserCfg{
			Timeout:   30 * time.Second,
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		}
	}

	return &Parser{cfg: cfg}
}

func (p *Parser) Parse(url *nurl.URL) (*Article, error) {
	if url == nil {
		return nil, fmt.Errorf("invalid URL")
	}

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	req.Header.Set("User-Agent", p.cfg.UserAgent)

	hclient := http.Client{
		Timeout: p.cfg.Timeout,
	}

	res, err := hclient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse html from response body: %w", err)
	}

	article, err := p.ParseDoc(doc, url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse doc: %w", err)
	}

	return article, nil
}

type Article struct {
	Title       string
	Description string
	Author      string
	PublishedAt *time.Time
	Poster      string
	Favicon     string
	Lang        string
	SourceURL   *nurl.URL
	Publisher   string
	HTML        *goquery.Document
	Text        string
	Length      int
	Images      []*body.Img
}

func (p *Parser) ParseDoc(doc *goquery.Document, baseURL *nurl.URL) (*Article, error) {
	if p.cfg.MaxItemsLen > 0 && doc.Length() > p.cfg.MaxItemsLen {
		return nil, fmt.Errorf("document is very big")
	}

	clone := goquery.CloneDocument(doc)

	meta := meta.Extract(clone.Selection, baseURL)
	body, err := body.Extract(clone, baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	return &Article{
		Title:       meta.Title,
		Description: meta.Description,
		Author:      meta.Author,
		PublishedAt: meta.PublishedAt,
		Poster:      meta.Poster,
		Favicon:     meta.Favicon,
		Lang:        meta.Lang,
		SourceURL:   meta.SourceURL,
		Publisher:   meta.Publisher,
		HTML:        body.HTML,
		Text:        body.Text,
		Length:      body.Len,
		Images:      body.Images,
	}, nil
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
