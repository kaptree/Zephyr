package repository

// pageSlice 对内存拼音过滤后的结果手动分页
func pageSlice[T any](items []T, page, pageSize int) []T {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		return items
	}
	offset := (page - 1) * pageSize
	if offset >= len(items) {
		return []T{}
	}
	end := offset + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}
