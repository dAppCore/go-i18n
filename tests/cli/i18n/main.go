package main

import (
	"fmt"
	"os"

	"dappco.re/go/i18n"
	log "dappco.re/go/log"
)

func main() {
	const op = "tests/cli/i18n.main"

	svc, err := i18n.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, log.E(op, "failed to initialise i18n service", err))
		os.Exit(1)
	}

	result := svc.T("test.key")
	fmt.Println(result)
	if result != "test.key" {
		fmt.Fprintf(os.Stderr, "unexpected translation result: want %q, got %q\n", "test.key", result)
		os.Exit(1)
	}
}
