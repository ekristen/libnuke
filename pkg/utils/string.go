package utils

func StringSliceContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func UnPtrBool(ptr *bool, def bool) bool {
	if ptr == nil {
		return def
	}
	return *ptr
}

func UnPtrString(ptr *string, def string) string {
	if ptr == nil {
		return def
	}
	return *ptr
}
