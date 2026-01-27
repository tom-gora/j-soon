// Package ical
package ical

import (
	"bufio"
	"io"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	l "github.com/tom-gora/JSON-from-iCal/internal/logger"
	"gitlab.com/metakeule/fmtdate"
)

func filterStream(r io.Reader, w *io.PipeWriter, matchList []string) {
	scanner := bufio.NewScanner(r)
	defer w.Close()
	for scanner.Scan() {
		line := scanner.Text()
		skip := false
		for _, m := range matchList {
			if len(m) > 0 && strings.Contains(line, m) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		w.Write([]byte(line + "\n"))
	}
	if err := scanner.Err(); err != nil {
		l.Log.Error.Println("reading input:", err)
	}
}

func getCalValueIfExists(e *ics.VEvent, p ics.ComponentProperty) string {
	prop := e.GetProperty(p)
	if prop != nil {
		return prop.Value
	}
	return ""
}

func dateStrToHuman(str string, templ string, now time.Time) string {
	if len(str) < 1 {
		return ""
	}
	t := StrToStructDate(str)
	if t.IsZero() {
		return ""
	}
	formattedDate := fmtdate.Format(templ, t)

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.AddDate(0, 0, 1)
	if t.Year() == today.Year() && t.Month() == today.Month() && t.Day() == today.Day() {
		formattedDate += " - TODAY ‼️"
	} else if t.Year() == tomorrow.Year() && t.Month() == tomorrow.Month() && t.Day() == tomorrow.Day() {
		formattedDate += " - TOMORROW ❗"
	}
	return formattedDate
}

func StrToStructDate(str string) time.Time {
	var t time.Time
	if len(str) == 8 {
		t, _ = time.ParseInLocation("20060102", str, time.Local)
	} else if len(str) == 15 {
		t, _ = time.ParseInLocation("20060102T150405", str, time.Local)
	} else if strings.HasSuffix(str, "Z") {
		t, _ = time.ParseInLocation("20060102T150405Z", str, time.UTC)
	} else {
		t, _ = time.ParseInLocation("20060102T150405", str, time.Local)
	}
	// still not good, go back to old format
	if t.IsZero() && len(str) > 0 {
		t, _ = time.ParseInLocation("20060102", str, time.Local)
	}
	return t
}

func zeroOutTimeFromDate(t time.Time) time.Time {
	t = t.Add(time.Duration(-t.Hour()) * time.Hour)
	t = t.Add(time.Duration(-t.Minute()) * time.Minute)
	t = t.Add(time.Duration(-t.Second()) * time.Second)
	t = t.Add(time.Duration(-t.Nanosecond()) * time.Nanosecond)
	return t
}

// shouldIncludeEvent determines if an event should be included based on the window
func shouldIncludeEvent(dtStart, dtEnd, windowStart, windowEnd time.Time) (keep bool, ongoing bool) {
	if dtEnd.Before(windowStart) || dtStart.After(windowEnd) {
		return false, false
	}
	if dtStart.Before(windowStart) {
		return true, true
	}
	return true, false
}

// parse ical event raw text to struct
func strToStructEvent(e *ics.VEvent, t string, now time.Time) CalendarEvent {
	uid := getCalValueIfExists(e, ics.ComponentPropertyUniqueId)
	start := getCalValueIfExists(e, ics.ComponentPropertyDtStart)
	actualEnd := getCalValueIfExists(e, ics.ComponentPropertyDtEnd)
	summary := getCalValueIfExists(e, ics.ComponentPropertySummary)
	location := getCalValueIfExists(e, ics.ComponentPropertyLocation)
	description := getCalValueIfExists(e, ics.ComponentPropertyDescription)

	st := StrToStructDate(start)
	et := StrToStructDate(actualEnd)
	us := st.Unix()
	ue := et.Unix()
	hours := et.Sub(st).Hours()

	subDay := hours < 24
	day := hours == 24
	multiDay := 24 < hours

	if len(actualEnd) < 1 {
		actualEnd = start
	}

	return CalendarEvent{
		UID:         uid,
		Start:       start,
		HumanStart:  dateStrToHuman(start, t, now),
		UnixStart:   us,
		End:         actualEnd,
		HumanEnd:    dateStrToHuman(actualEnd, t, now),
		UnixEnd:     ue,
		ActualEnd:   actualEnd,
		Summary:     summary,
		Location:    location,
		Description: description,
		Hours:       hours,
		SubDay:      subDay,
		Day:         day,
		MultiDay:    multiDay,
	}
}
