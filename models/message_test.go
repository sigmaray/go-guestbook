package models

import "testing"

func TestMessageDisplayAuthor(t *testing.T) {
	tests := []struct {
		name   string
		author string
		want   string
	}{
		{
			name:   "named author",
			author: "Alice",
			want:   "Alice",
		},
		{
			name:   "empty author",
			author: "",
			want:   "Anonymous",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Message{Author: tt.author}.DisplayAuthor()
			if got != tt.want {
				t.Fatalf("DisplayAuthor() = %q, want %q", got, tt.want)
			}
		})
	}
}
