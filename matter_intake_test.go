package main

import (
	"testing"
	"time"
)

func TestFollowUpStatusMovesSignedMatterPastDeadline(t *testing.T) {
	due := time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC)
	m := matter{Number: "MAT-1042", Deadline: due, SignedReady: true}
	got := followUpStatus(m, due.Add(time.Minute))
	if got != "deadline-follow-up" {
		t.Fatalf("expected deadline-follow-up, got %q", got)
	}
}
