package collection

func ToMap[K comparable, S, V any](s []S, conv func(S) (K, V)) map[K]V {
	m := make(map[K]V, len(s))
	for i := range s {
		k, v := conv(s[i])
		m[k] = v
	}

	return m
}

func ToMapP[K comparable, S, V any](s []S, conv func(*S) (K, V)) map[K]V {
	m := make(map[K]V, len(s))
	for i := range s {
		k, v := conv(&s[i])
		m[k] = v
	}

	return m
}

func Keys[K comparable, V any](m map[K]V) []K {
	res := make([]K, 0, len(m))
	for k := range m {
		res = append(res, k)
	}

	return res
}

func Values[K comparable, V any](m map[K]V) []V {
	res := make([]V, 0, len(m))
	for _, v := range m {
		res = append(res, v)
	}

	return res
}
