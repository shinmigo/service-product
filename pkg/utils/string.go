package utils

//首字母转大写
func FirstLitterToUpper(str string) string {
	if len(str) == 0 {
		return ""
	}

	list := []rune(str)
	if list[0] >= 97 && list[0] <= 122 {
		list[0] -= 32
	}

	return string(list)
}
