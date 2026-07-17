package msgops_test

import (
	"testing"

	"go-guestbook/msgops"
)

func TestWouldExceedMax(t *testing.T) {
	tests := []struct {
		name       string
		count      int64
		max        int
		additional int
		want       bool
	}{
		{name: "under cap", count: 2, max: 3, additional: 1, want: false},
		{name: "exactly at cap after add", count: 2, max: 3, additional: 1, want: false},
		{name: "at cap already", count: 3, max: 3, additional: 1, want: true},
		{name: "over cap", count: 4, max: 3, additional: 1, want: true},
		{name: "batch would exceed", count: 1, max: 3, additional: 3, want: true},
		{name: "batch fits", count: 1, max: 3, additional: 2, want: false},
		{name: "invalid max", count: 0, max: 0, additional: 1, want: true},
		{name: "zero additional treated as one", count: 0, max: 1, additional: 0, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := msgops.WouldExceedMax(tt.count, tt.max, tt.additional)
			if got != tt.want {
				t.Fatalf("WouldExceedMax(%d, %d, %d) = %v, want %v", tt.count, tt.max, tt.additional, got, tt.want)
			}
		})
	}
}
