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

func Exists(values []string, target string) bool {
	for _, value := range (values) {
		if value == target {
			return true
		}
	}
	return false
}
