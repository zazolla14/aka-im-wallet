package helper

func ChainString(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
