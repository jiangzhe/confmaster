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

type LinkedMap struct {
	m map[string]*Value
	ks []string
}

func NewLinkedMap(capacity int) *LinkedMap {
	return &LinkedMap{
		m: make(map[string]*Value, capacity),
		ks: make([]string, 0, capacity),
	}
}

func (lm *LinkedMap) Put(key string, value *Value) {
	_, exists := lm.m[key]
	if !exists {
		lm.ks = append(lm.ks, key)
	}
	lm.m[key] = value
}

func (lm *LinkedMap) Del(key string) {
	_, exists := lm.m[key]
	if exists {
		found := -1
		for idx, elem := range lm.ks {
			if elem == key {
				found = idx
				break
			}
		}
		if found != -1 {
			copy(lm.ks[found:], lm.ks[found+1:])
			lm.ks = lm.ks[:len(lm.ks)-1]
		}
	}
	delete(lm.m, key)
}

func (lm *LinkedMap) Keys() []string {
	keys := make([]string, len(lm.ks))
	for i, elem := range lm.ks {
		keys[i] = elem
	}
	return keys
}

func (lm *LinkedMap) Len() int {
	return len(lm.m)
}

func (lm *LinkedMap) Get(key string) *Value {
	return lm.m[key]
}
