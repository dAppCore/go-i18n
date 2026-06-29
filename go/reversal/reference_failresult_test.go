package reversal

import (
	"testing"

	"dappco.re/go"
)

// TestReferenceFailResult covers each branch of the reversal-package
// result-coercion helper: an already-failed Result passes through, an OK Result
// carrying an error is unwrapped, an OK Result with a non-error value is
// wrapped, a bare error is wrapped, and a plain value is stringified.
func TestReferenceFailResult(t *testing.T) {
	if got := failResult(core.Fail(core.NewError("boom"))); got.OK {
		t.Error("failResult(failed result) should stay failed")
	}
	if got := failResult(core.Ok(core.NewError("inner"))); got.OK {
		t.Error("failResult(ok-with-error) should fail")
	}
	if got := failResult(core.Ok("not-an-error")); got.OK {
		t.Error("failResult(ok-with-value) should fail")
	}
	if got := failResult(core.NewError("bare")); got.OK {
		t.Error("failResult(error) should fail")
	}
	if got := failResult(99); got.OK {
		t.Error("failResult(value) should fail")
	}
}
