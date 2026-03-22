module dappco.re/go/core/i18n/integration

go 1.25.5

require (
	dappco.re/go/core/i18n v0.0.0-00010101000000-000000000000
	forge.lthn.ai/core/go-inference v0.0.0
	forge.lthn.ai/core/go-mlx v0.0.0-00010101000000-000000000000
)

require golang.org/x/text v0.33.0 // indirect

replace (
	dappco.re/go/core/i18n => ../
	forge.lthn.ai/core/go-inference => ../../go-inference
	forge.lthn.ai/core/go-mlx => ../../go-mlx
)
