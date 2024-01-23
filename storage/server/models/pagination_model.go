package models

type Pagination struct {
	Cursor string `json:"cursor"`
	Limit  int64  `json:"limit"`
}

type PaginationResult struct {
	HasPrevious    bool   `json:"has_prev"`
	PreviousCursor string `json:"prev_cursor"`
	HasNext        bool   `json:"has_next"`
	NextCursor     string `json:"next_cursor"`
}
