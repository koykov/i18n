package i18n

type index map[string]int

func (i *index) set(key string, idx int) {
	(*i)[key] = idx
}

func (i index) get(key string) int {
	if i, ok := i[key]; ok {
		return i
	}
	return -1
}
