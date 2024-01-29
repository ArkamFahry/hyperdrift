package dtos

type Pagination struct {
	Cursor string `json:"cursor"`
	Limit  int32  `json:"limit"`
}

func (p *Pagination) SetDefaults() {
	if p.Limit == 0 {
		p.Limit = 10
	}
}

type PaginationResult struct {
	HasPrevious    bool   `json:"has_prev"`
	PreviousCursor string `json:"prev_cursor"`
	HasNext        bool   `json:"has_next"`
	NextCursor     string `json:"next_cursor"`
}
