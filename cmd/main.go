package main

import (
	"fmt"
	nurl "net/url"

	parser "github.com/raqiq-coder/articles-parser"
)

// const url = "https://habr.com/ru/companies/X5Tech/articles/1001330/"
// const url = "https://habr.com/ru/articles/1015700/"

// const url = "https://dev.to/ghostbuild/your-agent-can-think-it-cant-remember-5e1o"
// const url = "https://dev.to/allenarduino/creating-a-fully-functional-contact-form-with-react-and-formgrid-api-499m"
// const url = "https://vc.ru/ai/2835703-qwen3-5-9b-uncensored-reviz-neironki-bez-tsenzury"
// const url = "https://vc.ru/aviasales/2755911-statistika-puteshestviy-aviasales"
// const url = "https://vc.ru/education/2760692-kompetentsii-lidera-transformatsii"

const url = "https://www.computerra.ru/337572/kak-v-2000-h-skachivali-filmy-i-igry-na-kompyuter/"

func main() {
	parser := parser.NewParser(nil)

	parsed, err := nurl.Parse(url)
	if err != nil {
		fmt.Printf("ERROR: %v", err)
	}

	article, err := parser.Parse(parsed)
	if err != nil {
		fmt.Printf("ERROR: %v", err)
	}

	fmt.Println(article.Content.Html())
}
