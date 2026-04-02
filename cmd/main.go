package main

import (
	"fmt"

	nurl "net/url"

	parser "github.com/raqiq-coder/articles-parser"
)

const url = "https://habr.com/ru/companies/X5Tech/articles/1001330/"

// const url = "https://habr.com/ru/articles/1015700/"
// const url = "https://dev.to/ghostbuild/your-agent-can-think-it-cant-remember-5e1o"
// const url = "https://dev.to/allenarduino/creating-a-fully-functional-contact-form-with-react-and-formgrid-api-499m"
// const url = "https://vc.ru/ai/2835703-qwen3-5-9b-uncensored-reviz-neironki-bez-tsenzury"
// const url = "https://vc.ru/aviasales/2755911-statistika-puteshestviy-aviasales"
// const url = "https://vc.ru/education/2760692-kompetentsii-lidera-transformatsii"
// const url = "https://www.computerra.ru/337572/kak-v-2000-h-skachivali-filmy-i-igry-na-kompyuter/"
// const url = "https://www.joelonsoftware.com/2000/03/28/ndas-and-contracts-that-you-should-never-sign/" // без jsonld
// const url = "https://martinfowler.com/articles/reduce-friction-ai" // без jsonld
// const url = "https://easyperf.net/blog/2024/05/10/Thread-Count-Scaling-Part3" // нет ни og, ни twitter, ни jsonld
// const url = "https://travisdowns.github.io/blog/2020/07/06/concurrency-costs.html"
// const url = "https://shipilev.net/jvm/anatomy-quarks/12-native-memory-tracking/"
// const url = "https://fgiesen.wordpress.com/2025/05/21/oodle-2-9-14-and-intel-13th-14th-gen-cpus/"
// const url = "https://fsharpforfunandprofit.com/posts/mathematical-functions/"
// const url = "https://code.visualstudio.com/blogs/2026/02/05/multi-agent-development"

func main() {
	parsed, err := nurl.Parse(url)
	if err != nil {
		fmt.Printf("ERROR: %v", err)
	}

	parser := parser.NewParser(nil)
	article, err := parser.Parse(parsed)
	if err != nil {
		fmt.Printf("ERROR: %v", err)
	}

	article.PrintMeta()

	// fmt.Println(article.Content.Html())
}
