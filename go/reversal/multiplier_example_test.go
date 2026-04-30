package reversal

func ExampleNewMultiplier() {
	_ = NewMultiplier
}

func ExampleNewMultiplierForLang() {
	_ = NewMultiplierForLang
}

func ExampleMultiplier_Expand() {
	v := &Multiplier{}
	_ = v.Expand
}
