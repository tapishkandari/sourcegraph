package command

import "sort"

func orderedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

func intersperse(flag string, values []string) []string {
	interspersed := make([]string, 0, len(values))
	for _, v := range values {
		interspersed = append(interspersed, flag, v)
	}

	return interspersed
}

func flatten(values ...interface{}) []string {
	union := make([]string, 0, len(values))
	for _, value := range values {
		switch v := value.(type) {
		case string:
			union = append(union, v)
		case []string:
			union = append(union, v...)
		}
	}

	return union
}
