package util

func StringArrayContains(arr []string, item string) bool {
	for _, v := range arr {
		if v == item {
			return true
		}
	}
	return false
}

func RemoveFromSlice(arr []string, item string) []string {
	newArr := make([]string, 0)
	for _, v := range arr {
		if v != item {
			newArr = append(newArr, v)
		}
	}
	return newArr
}
