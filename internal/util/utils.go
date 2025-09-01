package util

func MapN[T, V any](ts []T, fn func(T) (V, error)) []V {
	result := make([]V, len(ts))
	i := 0
	for i < len(ts) {
		if v, err := fn(ts[i]); err == nil {
			result[i] = v
		}
		i++
	}

	return result[:i]
}

func Filter[T any](ts []T, fn func(T) bool) []T {
	result := []T{}
	for _, v := range ts {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

func Reduce[T, V any](ts []T, acc func(t T, v V) V, base V) V {
	for _, v := range ts {
		base = acc(v, base)
	}

	return base
}
