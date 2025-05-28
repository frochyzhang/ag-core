package reader

type Reader func(b []byte) (map[string]any, error)
