package body

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestBodyExtraction(t *testing.T) {
	cases := []struct {
		baseURL string
		name    string
	}{
		{
			baseURL: "https://habr.com/ru/companies/X5Tech/articles/1001330/",
			name:    "habr",
		},
		{
			baseURL: "https://dev.to/ghostbuild/your-agent-can-think-it-cant-remember-5e1o",
			name:    "dev",
		},
		{
			baseURL: "https://vc.ru/aviasales/2755911-statistika-puteshestviy-aviasales",
			name:    "vc",
		},
		{
			baseURL: "https://www.computerra.ru/337572/kak-v-2000-h-skachivali-filmy-i-igry-na-kompyuter/",
			name:    "computerra",
		},
		{
			baseURL: "https://www.joelonsoftware.com/2000/03/28/ndas-and-contracts-that-you-should-never-sign/",
			name:    "joelonsoftware",
		},
		{
			baseURL: "https://martinfowler.com/articles/reduce-friction-ai",
			name:    "martinfowler",
		},
		{
			baseURL: "https://easyperf.net/blog/2024/05/10/Thread-Count-Scaling-Part3",
			name:    "easyperf",
		},
		{
			baseURL: "https://travisdowns.github.io/blog/2020/07/06/concurrency-costs.html",
			name:    "travisdowns",
		},
		{
			baseURL: "https://fgiesen.wordpress.com/2025/05/21/oodle-2-9-14-and-intel-13th-14th-gen-cpus/",
			name:    "fgiesen",
		},
		{
			baseURL: "https://fsharpforfunandprofit.com/posts/mathematical-functions/",
			name:    "fsharpforfunandprofit",
		},
		{
			baseURL: "https://code.visualstudio.com/blogs/2026/02/05/multi-agent-development",
			name:    "code",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inputPath := fmt.Sprintf("testdata/%s/input.html", tc.name)
			inputContent, err := os.ReadFile(inputPath)
			if err != nil {
				t.Fatalf("failed to read input file %s: %v", inputPath, err)
			}

			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(inputContent))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}

			parsed, err := url.Parse(tc.baseURL)
			if err != nil {
				t.Fatalf("failed to parse baseURL: %v", err)
			}

			result, err := Extract(doc, parsed)
			if err != nil {
				t.Fatalf("failed to extract: %v", err)
			}

			var resHTML string
			if result.HTML != nil {
				resHTML, err = result.HTML.Html()
				if err != nil {
					t.Fatalf("failed to get HTML from result: %v", err)
				}
			}

			expectedPath := fmt.Sprintf("testdata/%s/want.html", tc.name)
			expectedContent, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("failed to read expected file %s: %v", expectedPath, err)
			}

			if strings.TrimSpace(resHTML) != strings.TrimSpace(string(expectedContent)) {
				t.Errorf("result not match with template")
			}
		})
	}
}
