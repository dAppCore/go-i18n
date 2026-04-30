package integration

import "dappco.re/go"

func valueFromResult[T any](r core.Result) (T, error) {
	var zero T
	if !r.OK {
		if err, ok := r.Value.(error); ok {
			return zero, err
		}
		return zero, core.NewError(r.Error())
	}
	v, ok := r.Value.(T)
	if !ok {
		return zero, core.NewError(core.Sprintf("unexpected result value %T", r.Value))
	}
	return v, nil
}
