package handlers

import (
	"strconv"
)

const (
	// adminListPageSize is how many rows admin list pages show per page.
	adminListPageSize = 10
)

// PaginationPageLink is one numbered page link in a paginated list.
type PaginationPageLink struct {
	Page   int
	URL    string
	Active bool
}

// PaginationView holds pagination state and navigation links for templates.
type PaginationView struct {
	Label      string
	Page       int
	TotalPages int
	Total      int64
	HasPrev    bool
	HasNext    bool
	PrevURL    string
	NextURL    string
	Pages      []PaginationPageLink
}

// parseQueryPage reads a one-based page number from a query string value.
// raw is the unparsed page query parameter; invalid or missing values become 1.
func parseQueryPage(raw string) int {
	page, err := strconv.Atoi(raw)
	if err != nil || page < 1 {
		return 1
	}
	return page
}

// totalPages returns the number of pages needed for total items at perPage size.
// total is the item count; perPage is the page size (values below 1 yield one page).
func totalPages(total int64, perPage int) int {
	if perPage < 1 {
		return 1
	}
	if total == 0 {
		return 1
	}
	return int((total + int64(perPage) - 1) / int64(perPage))
}

// clampPage limits page to a valid range for the given total and page size.
// page is the requested 1-based page; total is the item count; perPage is the page size.
func clampPage(page int, total int64, perPage int) int {
	if page < 1 {
		page = 1
	}
	maxPage := totalPages(total, perPage)
	if page > maxPage {
		return maxPage
	}
	return page
}

// pageOffset returns the SQL offset for a one-based page number.
// page is the 1-based page index; perPage is the page size.
func pageOffset(page, perPage int) int {
	return (page - 1) * perPage
}

// buildAdminListURL builds a paginated admin list URL using the page query param.
// path is the list path without a query string; page is the 1-based page number.
func buildAdminListURL(path string, page int) string {
	if page <= 1 {
		return path
	}
	return path + "?page=" + strconv.Itoa(page)
}

// buildPaginationView prepares template data for a paginated list section.
// total is the full item count; page and perPage describe the current window;
// label is used for the nav aria-label; urlForPage builds a URL for each page number.
func buildPaginationView(total int64, page, perPage int, label string, urlForPage func(page int) string) PaginationView {
	page = clampPage(page, total, perPage)
	pages := totalPages(total, perPage)

	view := PaginationView{
		Label:      label,
		Page:       page,
		TotalPages: pages,
		Total:      total,
		HasPrev:    page > 1,
		HasNext:    page < pages,
	}
	if view.HasPrev {
		view.PrevURL = urlForPage(page - 1)
	}
	if view.HasNext {
		view.NextURL = urlForPage(page + 1)
	}
	for p := 1; p <= pages; p++ {
		view.Pages = append(view.Pages, PaginationPageLink{
			Page:   p,
			URL:    urlForPage(p),
			Active: p == page,
		})
	}
	return view
}
