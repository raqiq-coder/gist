package main

import (
	"fmt"
	nurl "net/url"
	"os"
	"strings"
)

func writeExpectedHTMLToTestdata(url, html string) {
	parsed, err := nurl.Parse(url)
	if err != nil {
		fmt.Printf("ERROR: %v", err)
	}

	name := strings.Split(parsed.Host, ".")
	if name[0] == "www" {
		name[0] = name[1]
	}

	path := fmt.Sprintf("body/testdata/%s/want.html", name[0])

	bytes := []byte(html)
	err = os.WriteFile(path, bytes, 0644)
	if err != nil {
		fmt.Println("ERROR", err)
	}
}
