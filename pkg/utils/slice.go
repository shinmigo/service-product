package utils

//是否uint64 slice中存在
func InSliceUint64(v uint64, newSlice []uint64) bool {
	for k := range newSlice {
		if newSlice[k] == v {
			return true
		}
	}
	return false
}

//是否string slice中存在
func InSliceString(v string, newSlice []string) bool {
	for k := range newSlice {
		if newSlice[k] == v {
			return true
		}
	}
	return false
}

//去重
func SliceUniqueString(sliceList []string) []string {
	result := sliceList[:0]
	for k := range sliceList {
		if !InSliceString(sliceList[k], result) {
			result = append(result, sliceList[k])
		}
	}

	return result
}

//返回slice, slice中的值是slice1在slice2中不存在的 返回差集
func SliceDiff(slice1, slice2 []string) (diffSlice []string) {
	for k := range slice1 {
		if !InSliceString(slice1[k], slice2) {
			diffSlice = append(diffSlice, slice1[k])
		}
	}
	return
}

//返回slice  slice中的值是slice1和slice2中同时存在的，返回交集
func SliceIntersect(slice1, slice2 []string) (intersectSlice []string) {
	for k := range slice1 {
		if InSliceString(slice1[k], slice2) {
			intersectSlice = append(intersectSlice, slice1[k])
		}
	}
	return
}

func SliceDeleteString(val string, sliceList []string) []string {
	if len(sliceList) == 0 {
		return []string{}
	}

	index := 0
	endIndex := len(sliceList) - 1
	result := make([]string, 0, len(sliceList))
	for k := range sliceList {
		if sliceList[k] == val {
			result = append(result, sliceList[index:k]...)
			index = k + 1
		} else if k == endIndex {
			result = append(result, sliceList[index:endIndex+1]...)
		}
	}

	return result
}

func SliceDeleteUint64(val uint64, sliceList []uint64) []uint64 {
	if len(sliceList) == 0 {
		return []uint64{}
	}

	index := 0
	endIndex := len(sliceList) - 1
	result := make([]uint64, 0, len(sliceList))
	for k := range sliceList {
		if sliceList[k] == val {
			result = append(result, sliceList[index:k]...)
			index = k + 1
		} else if k == endIndex {
			result = append(result, sliceList[index:endIndex+1]...)
		}
	}

	return result
}
