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
	"github.com/tom-gora/JSON-from-iCal/internal/common"
	l "github.com/tom-gora/JSON-from-iCal/internal/logger"
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

	if common.IsValidFile(uri) {
		return os.Open(common.PathExpandTilde(uri))
	}

	return nil, fmt.Errorf("invalid source: %s", uri)
}

func ProcessSourceToStruct(r io.Reader, today time.Time, t string, u int) []CalendarEvent {
	var events []CalendarEvent

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

	for _, e := range cal.Events() {
		event := strToStructEvent(e, t, today)
		dtStart := StrToStructDate(event.Start)
		dtEnd := StrToStructDate(event.End)
		if dtEnd.IsZero() {
			dtEnd = dtStart
		}

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
	return events
}
