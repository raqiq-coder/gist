package parser

import (
	"fmt"
	"net/http"
	nurl "net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Parser struct {
	cfg *ParserCfg

	doc     *goquery.Selection
	baseURL *nurl.URL
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
	Images      []*ImgMeta
}

func (p *Parser) ParseDoc(doc *goquery.Document, baseURL *nurl.URL) (*Article, error) {
	if p.cfg.MaxItemsLen > 0 && doc.Length() > p.cfg.MaxItemsLen {
		return nil, fmt.Errorf("document is very big")
	}

	clone := goquery.CloneDocument(doc)

	p.baseURL = baseURL
	p.doc = clone.Selection

	meta := p.extractMeta()
	con, err := p.extractContent()
	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	return &Article{
		Title:       meta.title,
		Description: meta.description,
		Author:      meta.author,
		PublishedAt: meta.publishedAt,
		Poster:      meta.poster,
		Favicon:     meta.favicon,
		Lang:        meta.lang,
		SourceURL:   meta.sourceURL,
		Publisher:   meta.publisher,
		HTML:        con.html,
		Text:        con.text,
		Length:      con.len,
		Images:      con.images,
	}, nil
}
