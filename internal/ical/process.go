// Package ical
package ical

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	fu "github.com/tom-gora/j-soon/internal/fileutil"
	l "github.com/tom-gora/j-soon/internal/logger"
)

type CalendarEvent struct {
	UID         string  `json:"UID"`
	Start       string  `json:"Start"`
	HumanStart  string  `json:"HumanStart"`
	UnixStart   int64   `json:"UnixStart"`
	End         string  `json:"End"`
	HumanEnd    string  `json:"HumanEnd"`
	UnixEnd     int64   `json:"UnixEnd"`
	ActualEnd   string  `json:"ActualEnd"`
	Summary     string  `json:"Summary"`
	Location    string  `json:"Location"`
	Description string  `json:"Description"`
	Hours       float64 `json:"Hours"`
	SubDay      bool    `json:"SubDay"`
	Day         bool    `json:"Day"`
	MultiDay    bool    `json:"MultiDay"`
	Ongoing     bool    `json:"Ongoing"`
}

func FetchSource(uri string) (io.ReadCloser, error) {
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		c := &http.Client{}
		req, err := http.NewRequest("GET", uri, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0")

		resp, err := c.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("bad status: %s", resp.Status)
		}
		return resp.Body, nil
	}

	expandedPath := fu.PathExpandTilde(uri)
	if info, err := os.Stat(expandedPath); err == nil && !info.IsDir() {
		return os.Open(expandedPath)
	}

	return nil, fmt.Errorf("invalid source: %s", uri)
}

func ProcessSourceToStruct(r io.Reader, today time.Time, t string, u int, markers map[string]string) []CalendarEvent {
	events := []CalendarEvent{}

	// Apply X-APPLE- filtering
	pr, pw := io.Pipe()
	go filterStream(r, pw, []string{"X-APPLE-"})

	cal, err := ics.ParseCalendar(pr)
	if err != nil {
		l.Log.Error.Println("parsing calendar:", err)
		return events
	}

	windowStart := zeroOutTimeFromDate(today)
	windowEnd := windowStart.AddDate(0, 0, u).Add(24 * time.Hour).Add(-time.Second)

	totalEvents := len(cal.Events())
	for _, e := range cal.Events() {
		event := strToStructEvent(e, t, today, markers)
		dtStart := time.Unix(event.UnixStart, 0)
		dtEnd := time.Unix(event.UnixEnd, 0)

		keep, ongoing := shouldIncludeEvent(dtStart, dtEnd, windowStart, windowEnd)
		if !keep {
			continue
		}

		if ongoing {
			event.Ongoing = true
			event.Description = "Ongoing\n\n" + event.Description
		}

		events = append(events, event)
	}

	l.Log.Debug.Printf("Processed calendar: found %d events, kept %d within window", totalEvents, len(events))
	return events
}
