package main

import (
	"strings"
	"testing"
	"time"
)

func TestRawToStructDate(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		{"20260122", time.Date(2026, 1, 22, 0, 0, 0, 0, time.Local)},
		{"20260122T140000", time.Date(2026, 1, 22, 14, 0, 0, 0, time.Local)},
		{"20260122T140000Z", time.Date(2026, 1, 22, 14, 0, 0, 0, time.UTC)},
		{"", time.Time{}},
		{"invalid", time.Time{}},
	}

	for _, tt := range tests {
		got := rawToStructDate(tt.input)
		if !got.Equal(tt.expected) {
			t.Errorf("rawToStructDate(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestZeroOutTimeFromDate(t *testing.T) {
	input := time.Date(2026, 1, 22, 14, 30, 15, 123456789, time.Local)
	expected := time.Date(2026, 1, 22, 0, 0, 0, 0, time.Local)
	got := zeroOutTimeFromDate(input)
	if !got.Equal(expected) {
		t.Errorf("zeroOutTimeFromDate(%v) = %v, want %v", input, got, expected)
	}
}

func TestDateStrToHuman(t *testing.T) {
	now := time.Date(2026, 1, 22, 12, 0, 0, 0, time.Local)
	tests := []struct {
		input    string
		contains string
	}{
		{"20260122T140000", "TODAY ‼️"},
		{"20260123T100000", "TOMORROW ❗"},
		{"20260124T100000", "[ SAT ] 24 Jan 2026 @ 10:00"},
	}

	for _, tt := range tests {
		got := dateStrToHuman(tt.input, now)
		if !strings.Contains(got, tt.contains) {
			t.Errorf("dateStrToHuman(%q, now) = %q, should contain %q", tt.input, got, tt.contains)
		}
	}
}

func TestShouldIncludeEvent(t *testing.T) {
	windowStart := time.Date(2026, 1, 22, 0, 0, 0, 0, time.Local)
	windowEnd := windowStart.Add(24 * time.Hour).Add(-time.Second) // End of today

	tests := []struct {
		name    string
		start   time.Time
		end     time.Time
		keep    bool
		ongoing bool
	}{
		{
			name:    "Event in window",
			start:   windowStart.Add(2 * time.Hour),
			end:     windowStart.Add(3 * time.Hour),
			keep:    true,
			ongoing: false,
		},
		{
			name:    "Event before window",
			start:   windowStart.Add(-5 * time.Hour),
			end:     windowStart.Add(-2 * time.Hour),
			keep:    false,
			ongoing: false,
		},
		{
			name:    "Event after window",
			start:   windowEnd.Add(1 * time.Hour),
			end:     windowEnd.Add(2 * time.Hour),
			keep:    false,
			ongoing: false,
		},
		{
			name:    "Ongoing event",
			start:   windowStart.Add(-2 * time.Hour),
			end:     windowStart.Add(2 * time.Hour),
			keep:    true,
			ongoing: true,
		},
		{
			name:    "Boundary check start",
			start:   windowStart,
			end:     windowStart.Add(1 * time.Hour),
			keep:    true,
			ongoing: false,
		},
		{
			name:    "Boundary check end",
			start:   windowEnd.Add(-1 * time.Hour),
			end:     windowEnd,
			keep:    true,
			ongoing: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keep, ongoing := shouldIncludeEvent(tt.start, tt.end, windowStart, windowEnd)
			if keep != tt.keep || ongoing != tt.ongoing {
				t.Errorf("shouldIncludeEvent() = (%v, %v), want (%v, %v)", keep, ongoing, tt.keep, tt.ongoing)
			}
		})
	}
}
