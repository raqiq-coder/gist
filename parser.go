package parser

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
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
	if cfg != nil {
		return &Parser{
			cfg,
		}
	}

	return &Parser{
		cfg: &ParserCfg{
			Timeout:   30 * time.Second,
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		},
	}

}

func (p *Parser) Parse(url *url.URL) (*Article, error) {
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

func (p *Parser) ParseDoc(doc *goquery.Document, baseURL *url.URL) (*Article, error) {
	if p.cfg.MaxItemsLen > 0 && doc.Length() > p.cfg.MaxItemsLen {
		return nil, fmt.Errorf("document is very big")
	}

	clone := goquery.CloneDocument(doc)
	article := &Article{
		doc:     clone.Selection,
		baseURL: baseURL,
	}

	article.getMeta()

	if err := article.getArticle(); err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	return article, nil
}
