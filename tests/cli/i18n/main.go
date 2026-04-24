package main

import (
	"fmt"
	"os"

	"dappco.re/go/i18n"
)

func main() {
	svc, err := i18n.New()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	result := svc.T("test.key")
	fmt.Println(result)
	if result != "test.key" {
		os.Exit(1)
	}
}
