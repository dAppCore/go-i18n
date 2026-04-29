module dappco.re/go/i18n

go 1.26.0

require golang.org/x/text v0.36.0

require (
	dappco.re/go v0.9.0
	dappco.re/go/inference v0.8.0-alpha.1
	dappco.re/go/log v0.8.0-alpha.1
)

replace (
	dappco.re/go => ../go
	dappco.re/go/inference => ../go-inference
	dappco.re/go/log => ../go-log
)
