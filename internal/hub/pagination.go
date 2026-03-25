package hub

// PageItem represents a single item in the pagination control bar.
type PageItem struct {
	Number   int  // Page number (0 for ellipsis sentinel)
	IsActive bool // Whether this is the current page
	IsGap    bool // Whether this represents an ellipsis ("...") gap
}

// PaginationData holds pre-computed pagination metadata for templates.
type PaginationData struct {
	CurrentPage int
	TotalPages  int
	TotalItems  int
	Pages       []PageItem
	HasPrev     bool
	HasNext     bool
	PrevPage    int
	NextPage    int
}

// calcPagination computes pagination metadata for the given total items,
// requested page number, and page size. Out-of-range page values are clamped.
func calcPagination(totalItems, page, pageSize int) PaginationData {
	if totalItems <= 0 || pageSize <= 0 {
		return PaginationData{
			CurrentPage: 1,
			TotalPages:  0,
			TotalItems:  totalItems,
		}
	}

	totalPages := (totalItems + pageSize - 1) / pageSize

	// Clamp page to valid range.
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	pd := PaginationData{
		CurrentPage: page,
		TotalPages:  totalPages,
		TotalItems:  totalItems,
		HasPrev:     page > 1,
		HasNext:     page < totalPages,
	}
	if pd.HasPrev {
		pd.PrevPage = page - 1
	}
	if pd.HasNext {
		pd.NextPage = page + 1
	}

	pd.Pages = buildPageItems(page, totalPages)
	return pd
}

// buildPageItems generates the list of page items for rendering.
// Shows all pages when totalPages ≤ 7, otherwise uses ellipsis truncation.
func buildPageItems(current, totalPages int) []PageItem {
	if totalPages <= 1 {
		return nil
	}

	if totalPages <= 7 {
		items := make([]PageItem, totalPages)
		for i := range items {
			items[i] = PageItem{Number: i + 1, IsActive: i+1 == current}
		}
		return items
	}

	// Ellipsis strategy: 1 ... (current-1) current (current+1) ... last
	var items []PageItem
	add := func(n int) {
		items = append(items, PageItem{Number: n, IsActive: n == current})
	}
	gap := func() {
		items = append(items, PageItem{IsGap: true})
	}

	add(1)

	if current > 3 {
		gap()
	}

	for p := current - 1; p <= current+1; p++ {
		if p > 1 && p < totalPages {
			add(p)
		}
	}

	if current < totalPages-2 {
		gap()
	}

	add(totalPages)

	return items
}
