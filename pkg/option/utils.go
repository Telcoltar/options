package option

func SetNotNil[T any, B any](o OptionGet[T], targetFunc func(v T) B) {
	if o.HasValue() {
		targetFunc(o.Get())
	}
}
