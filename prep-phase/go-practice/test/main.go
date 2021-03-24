package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	resp, err := http.Get("https://golang.org/cmd/go/#hdr-Download_and_install_packages_and_dependencies")
	if err != nil {
		log.Fatalf("http.Get => %v", err.Error())
	}

	finalURL := resp.Request.URL.String()

	fmt.Println(finalURL)

}
