package utils

func IndexOf(slice []string, item string) int {
	for idx, o := range slice {
		if o == item {
			return idx
		}
	}

	return -1
}

func RemoveFromSlice(slice []string, item string) []string {
	idx := IndexOf(slice, item)
	return append(slice[:idx], slice[idx+1:]...)
}
