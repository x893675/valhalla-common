package ptr

type BasicType interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~complex64 | ~complex128 |
		~string | ~bool
}

func To[T BasicType](v T) *T {
	return &v
}

func From[T BasicType](v *T) T {
	if v == nil {
		var zero T
		return zero
	}
	return *v
}

func SlicePtr[T BasicType](vs []T) []*T {
	ps := make([]*T, len(vs))
	for i, v := range vs {
		vv := v
		ps[i] = &vv
	}
	return ps
}

// MapPtr returns a map of pointers from the values passed in.
func MapPtr[T BasicType](vs map[string]T) map[string]*T {
	ps := make(map[string]*T, len(vs))
	for k, v := range vs {
		vv := v
		ps[k] = &vv
	}
	return ps
}
