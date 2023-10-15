//go:generate go run -tags generate sync.go
//go:build generate
// +build generate

package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"go.pitz.tech/em/internal/project/scaffold/licenses"
)

func fetch(address string) (io.ReadCloser, error) {
	resp, err := http.Get(address)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func handleLicense(license licenses.License) error {
	target := "https://raw.githubusercontent.com/licenses/license-templates/master/templates/" + license.TemplateName + ".txt"
	destination := license.TemplateName + ".txt"

	handle, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer handle.Close()

	reader, err := fetch(target)
	if err != nil {
		return err
	}
	defer reader.Close()

	_, err = io.Copy(handle, reader)
	return err
}

func main() {
	for _, license := range licenses.BuiltIn {
		err := handleLicense(license)
		if err != nil {
			log.Printf("[%s] failed %v", license.Identifier, err)
		}
	}
}
