package reversal

func ExampleBuildReferences() {
	_ = BuildReferences
}

func ExampleReferenceSet_Compare() {
	v := &ReferenceSet{}
	_ = v.Compare
}

func ExampleReferenceSet_Classify() {
	v := &ReferenceSet{}
	_ = v.Classify
}

func ExampleReferenceSet_DomainNames() {
	v := &ReferenceSet{}
	_ = v.DomainNames
}
