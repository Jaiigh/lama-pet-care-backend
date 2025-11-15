package utils

func CalDefaultOffsetEnd(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 5
	}
	offset := (page - 1) * limit
	//return offset, end
	return offset, offset + limit
}
