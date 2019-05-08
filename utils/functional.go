package utils

func Filter(values []string, predicate func(string) bool) []string {
	result := make([]string, 0)
	for _, value := range (values) {
		if predicate(value) {
			result = append(result, value)
		}
	}
	return result
}
