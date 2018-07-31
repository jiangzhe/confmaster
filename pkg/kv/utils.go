package kv

type stringLenSorter struct {
	arr []string
}

func (sls *stringLenSorter) Len() int {
	return len(sls.arr)
}

func (sls *stringLenSorter) Less(i, j int) bool {
	return len(sls.arr[i]) < len(sls.arr[j])
}

func (sls *stringLenSorter) Swap(i, j int) {
	sls.arr[i], sls.arr[j] = sls.arr[j], sls.arr[i]
}