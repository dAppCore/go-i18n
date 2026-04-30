package main

import (
	"dappco.re/go"
	"dappco.re/go/i18n"
	golog "dappco.re/go/log"
)

func main() {
	const op = "tests/cli/i18n.main"

	r := i18n.New()
	if !r.OK {
		core.Print(core.Stderr(), "%v", golog.E(op, "failed to initialise i18n service", core.NewError(r.Error())))
		core.Exit(1)
	}
	svc := r.Value.(*i18n.Service)

	result := svc.T("test.key")
	core.Println(result)
	if result != "test.key" {
		core.Print(core.Stderr(), "unexpected translation result: want %q, got %q", "test.key", result)
		core.Exit(1)
	}
}
