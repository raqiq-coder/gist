package main

import (
	"fmt"
	"io"
	"net/http"
	nurl "net/url"
	"os"
	"strings"
)

func writeFixtureHTMLToTestdata(url string) {
	parsed, err := nurl.Parse(url)
	if err != nil {
		fmt.Printf("ERROR: %v", err)
	}

	client := http.Client{}

	name := strings.Split(parsed.Host, ".")
	if name[0] == "www" {
		name[0] = name[1]
	}

	req, _ := http.NewRequest("GET", parsed.String(), nil)
	res, _ := client.Do(req)

	if res.StatusCode == http.StatusOK {
		path := fmt.Sprintf("body/testdata/%s/input.html", name[0])

		os.Mkdir(fmt.Sprintf("body/testdata/%s", name[0]), 0755)

		bytes, _ := io.ReadAll(res.Body)
		err = os.WriteFile(path, bytes, 0644)
		if err != nil {
			fmt.Println("ERROR", err)
		}
	}
}
