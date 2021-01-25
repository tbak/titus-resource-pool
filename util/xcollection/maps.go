package xcollection

func SetOfStringList(stringList []string) map[string]bool {
	result := map[string]bool{}
	for _, item := range stringList {
		result[item] = true
	}
	return result
}
