package handlers

import "testing"

// TestTotalPages checks page-count math for empty, partial, and exact pages.
func TestTotalPages(t *testing.T) {
	tests := []struct {
		name    string
		total   int64
		perPage int
		want    int
	}{
		{name: "empty", total: 0, perPage: 10, want: 1},
		{name: "exact page", total: 20, perPage: 10, want: 2},
		{name: "partial page", total: 21, perPage: 10, want: 3},
		{name: "invalid per page", total: 5, perPage: 0, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := totalPages(tt.total, tt.perPage); got != tt.want {
				t.Fatalf("totalPages() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestClampPage checks that out-of-range page numbers are corrected.
func TestClampPage(t *testing.T) {
	tests := []struct {
		name    string
		page    int
		total   int64
		perPage int
		want    int
	}{
		{name: "below one", page: 0, total: 25, perPage: 10, want: 1},
		{name: "within range", page: 2, total: 25, perPage: 10, want: 2},
		{name: "above max", page: 9, total: 25, perPage: 10, want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clampPage(tt.page, tt.total, tt.perPage); got != tt.want {
				t.Fatalf("clampPage() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestBuildAdminListURL checks that page 1 omits the query string.
func TestBuildAdminListURL(t *testing.T) {
	if got := buildAdminListURL("/admin/messages", 1); got != "/admin/messages" {
		t.Fatalf("page 1 URL = %q, want /admin/messages", got)
	}
	if got := buildAdminListURL("/admin/users", 3); got != "/admin/users?page=3" {
		t.Fatalf("page 3 URL = %q, want /admin/users?page=3", got)
	}
}
