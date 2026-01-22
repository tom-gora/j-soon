package main

/* *****************************************************************************
DISCLAIMER: THIS PROGRAM IS BASED OF OFF:
https://github.com/jceaser/readical
By: thomas.cherry@gmail.com
---
This was heavily transofrmed/reduced and mostly served as reference of how ical processing is done

Significantly simplified and repurposed to write events as json objects rather than markup files
MODIFICATIONS by goratomasz@outlook.com

Use case: Backend to produce local system notifications from declared ical calendar files
(local and remote) with events upcoming in a given range of days.

Intended for small scope of logic and personal use.

List of modifications made to the fork:

- Write output data as JSON file if flag provided, otherwise without -f stdout is used
- Better flags system with long and shorthand flags.
- Better usage print.
- Remove timeranging logic with "out" and "after" values that were overkill for the usecase.
- Introduce single --upcoming-days flag to define how far out from today to look for events.
- Remove any markup processing and formatting. Marshal and write json string
- Remove control of used timezone. Again, local only use so grabbing os timezone is sufficient.
- Lasting events that come up "ongoing" as of today will be included and will have .Description "Ongoing".
- Script friendly as content of ics file can be piped in,
	or a -c flag to a config file listing local and remote calendars can be passed
- Tidied up codebase to remove missplelings, naming conventions I found unclear or unnecesarily verbose,
	fix case of names that was all over the place and made it all adhere to go recommended standards

***************************************************************************** */

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/arran4/golang-ical"
	"github.com/fatih/color"
	"github.com/rodaine/table"
)

const (
	AppBy      = "JSON-from-iCal by github.com/tom-gora"
	AppVersion = "1.0.0"
)

type logType struct {
	Report *log.Logger
	Error  *log.Logger
	Warn   *log.Logger
	Info   *log.Logger
	Debug  *log.Logger
	Stats  *log.Logger
}

var Log logType

func init() {
	file := os.Stderr
	settings := log.Ldate | log.Ltime | log.Lshortfile
	Log = logType{
		Report: log.New(file, "REPORT: ", settings),
		Error:  log.New(file, "ERROR: ", settings),
		Warn:   log.New(file, "WARNING: ", settings),
		Info:   log.New(io.Discard, "INFO: ", settings),
		Debug:  log.New(io.Discard, "DEBUG: ", settings),
		Stats:  log.New(io.Discard, "", settings),
	}
}

func (l *logType) EnableInfo() {
	l.Info.SetOutput(os.Stderr)
}

func (l *logType) EnableDebug() {
	l.Debug.SetOutput(os.Stderr)
}

func (l *logType) EnableStats(file *os.File) {
	l.Stats.SetOutput(file)
}

type exitCode int

const (
	ExitNorm exitCode = iota
	ExitVer
)

func (ec exitCode) int() int {
	return int(ec)
}

type execConf struct {
	CalendarsConfigPath string
	TargetFile          string
	Limit               int
	Verbose             bool
	UpcomingDays        int
}

type calendarEvent struct {
	UID         string  `json:"UID"`
	Start       string  `json:"Start"`
	HumanStart  string  `json:"HumanStart"`
	End         string  `json:"End"`
	HumanEnd    string  `json:"HumanEnd"`
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
		Log.Error.Println("reading input:", err)
	}
}

func getIfExists(e *ics.VEvent, p ics.ComponentProperty) string {
	prop := e.GetProperty(p)
	if prop != nil {
		return prop.Value
	}
	return ""
}

// parse ical event raw text to struct
func rawToStructEvent(e *ics.VEvent, now time.Time) calendarEvent {
	uid := getIfExists(e, ics.ComponentPropertyUniqueId)
	start := getIfExists(e, ics.ComponentPropertyDtStart)
	actualEnd := getIfExists(e, ics.ComponentPropertyDtEnd)
	summary := getIfExists(e, ics.ComponentPropertySummary)
	location := getIfExists(e, ics.ComponentPropertyLocation)
	description := getIfExists(e, ics.ComponentPropertyDescription)

	st := rawToStructDate(start)
	et := rawToStructDate(actualEnd)
	hours := et.Sub(st).Hours()

	subDay := hours < 24
	day := hours == 24
	multiDay := 24 < hours

	if len(actualEnd) < 1 {
		actualEnd = start
	}

	return calendarEvent{
		UID:         uid,
		Start:       start,
		HumanStart:  dateStrToHuman(start, now),
		End:         actualEnd,
		HumanEnd:    dateStrToHuman(actualEnd, now),
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

func rawToStructDate(str string) time.Time {
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

// format for notification - hardcoded to something akin:
// [ FRI ] 09 Jul 2021 @ 10:00 - TOMORROW !
func dateStrToHuman(str string, now time.Time) string {
	if len(str) < 1 {
		return ""
	}
	t := rawToStructDate(str)
	if t.IsZero() {
		return ""
	}
	day := strings.ToUpper(t.Format("Mon"))
	rest := t.Format("02 Jan 2006 @ 15:04")

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.AddDate(0, 0, 1)
	if t.Year() == today.Year() && t.Month() == today.Month() && t.Day() == today.Day() {
		rest += " - TODAY ‼️"
	} else if t.Year() == tomorrow.Year() && t.Month() == tomorrow.Month() && t.Day() == tomorrow.Day() {
		rest += " - TOMORROW ❗"
	}
	return fmt.Sprintf("[ %s ] %s", day, rest)
}

func initWithConf() execConf {
	conf := execConf{
		CalendarsConfigPath: "",
		TargetFile:          "",
		Limit:               0,
		Verbose:             false,
		UpcomingDays:        7,
	}
	return conf
}

func expandTilde(p string) string {
	if strings.HasPrefix(p, "~/") {
		dir, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		p = filepath.Join(dir, p[2:])
	}
	return p
}

func isValidFile(p string) bool {
	info, err := os.Stat(expandTilde(p))
	return err == nil && !info.IsDir()
}

func fetchSource(uri string) (io.ReadCloser, error) {
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

	if isValidFile(uri) {
		return os.Open(expandTilde(uri))
	}

	return nil, fmt.Errorf("invalid source: %s", uri)
}

func parseCalendarsConfig(p string) ([]string, error) {
	file, err := os.Open(expandTilde(p))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var uris []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Whitelist: Must be a URL or a valid local file
		isURL := strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://")
		if isURL || isValidFile(line) {
			uris = append(uris, expandTilde(line))
		}
	}
	return uris, scanner.Err()
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

func processSource(r io.Reader, today time.Time, c execConf) []calendarEvent {
	var events []calendarEvent

	// Apply X-APPLE- filtering
	pr, pw := io.Pipe()
	go filterStream(r, pw, []string{"X-APPLE-"})

	cal, err := ics.ParseCalendar(pr)
	if err != nil {
		Log.Error.Println("parsing calendar:", err)
		return events
	}

	windowStart := zeroOutTimeFromDate(today)
	windowEnd := windowStart.AddDate(0, 0, c.UpcomingDays).Add(24 * time.Hour).Add(-time.Second)

	for _, e := range cal.Events() {
		event := rawToStructEvent(e, today)
		dtStart := rawToStructDate(event.Start)
		dtEnd := rawToStructDate(event.End)
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

func resolveOutputPath(p string) string {
	if p != "" {
		return p
	}

	// Priority Logic
	xdgCache := os.Getenv("XDG_CACHE_HOME")
	if xdgCache != "" {
		return filepath.Join(xdgCache, "event-notifications", "out.json")
	}

	home, err := os.UserHomeDir()
	if err == nil {
		return filepath.Join(home, ".cache", "event-notifications", "out.json")
	}

	return filepath.Join(".", "out", "out.json")
}

func cleanOldJSONFiles(dir string) {
	files, err := filepath.Glob(filepath.Join(dir, "out.json"))
	if err != nil {
		return
	}
	for _, f := range files {
		os.Remove(f)
	}
}

func safeWrite(p string, data []byte) error {
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Clean up old JSON files in the destination directory to avoid mixing with old multi-file output
	cleanOldJSONFiles(dir)

	info, err := os.Stat(p)
	if err == nil {
		if info.IsDir() {
			return fmt.Errorf("target path is a directory: %s", p)
		}
	}

	return os.WriteFile(p, data, 0o644)
}

func main() {
	conf := initWithConf()

	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "\n%s\n", color.New(color.FgCyan, color.Bold).Sprint(AppBy))
		fmt.Fprintf(out, "Version: %s\n\n", color.New(color.FgYellow).Sprint(AppVersion))
		fmt.Fprintf(out, "%s\n", color.New(color.FgHiWhite, color.Underline).Sprint("Usage:"))
		fmt.Fprintf(out, "  jfi [flags]\n\n")

		headerFmt := color.New(color.FgHiGreen, color.Bold).SprintfFunc()
		columnFmt := color.New(color.FgHiWhite).SprintfFunc()

		tbl := table.New("Flag", "Short", "Default", "Description")
		tbl.WithWriter(out)
		tbl.WithHeaderFormatter(headerFmt)
		tbl.WithFirstColumnFormatter(columnFmt)

		flags := []struct {
			long, short, def, desc string
		}{
			{"upcoming-days", "u", "7", "Number of upcoming days to include"},
			{"limit", "l", "0", "Max number of events to process (0=unlimited)"},
			{"file", "f", "stdout", "Output file (empty for priority logic)"},
			{"config", "c", "", "Path to calendars.conf"},
			{"verbose", "v", "false", "Enable verbose logging"},
			{"version", "V", "false", "Report version info"},
		}

		for _, f := range flags {
			tbl.AddRow("--"+f.long, "-"+f.short, f.def, f.desc)
		}

		tbl.Print()
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Priority Logic for --file \"\":")
		fmt.Fprintln(out, "  1. $XDG_CACHE_HOME/event-notifications/out.json")
		fmt.Fprintln(out, "  2. $HOME/.cache/event-notifications/out.json")
		fmt.Fprintln(out, "  3. ./out/out.json")
	}

	var (
		version  bool
		verbose  bool
		filePath string
		fileSet  bool
		confPath string
		limit    int
		upcoming int
	)

	flag.BoolVar(&version, "version", false, "")
	flag.BoolVar(&version, "V", false, "")
	flag.BoolVar(&verbose, "verbose", false, "")
	flag.BoolVar(&verbose, "v", false, "")
	flag.IntVar(&limit, "limit", 0, "")
	flag.IntVar(&limit, "l", 0, "")
	flag.IntVar(&upcoming, "upcoming-days", 7, "")
	flag.IntVar(&upcoming, "u", 7, "")
	flag.StringVar(&confPath, "config", "", "")
	flag.StringVar(&confPath, "c", "", "")

	// Custom flag parsing for -f to detect if it's set at all
	flag.Func("file", "Output file", func(s string) error {
		filePath = s
		fileSet = true
		return nil
	})
	flag.Func("f", "Output file", func(s string) error {
		filePath = s
		fileSet = true
		return nil
	})

	flag.Parse()

	if version {
		fmt.Printf("%s\nVersion %s\n", AppBy, AppVersion)
		os.Exit(ExitNorm.int())
	}

	conf.Limit = limit
	conf.UpcomingDays = upcoming
	conf.Verbose = verbose
	conf.CalendarsConfigPath = confPath

	if verbose {
		Log.EnableInfo()
		Log.EnableDebug()
	}

	allEvents := []calendarEvent{}
	now := time.Now()

	if conf.CalendarsConfigPath != "" {
		uris, err := parseCalendarsConfig(conf.CalendarsConfigPath)
		if err != nil {
			Log.Error.Fatalf("failed to parse config: %v", err)
		}

		for _, uri := range uris {
			Log.Info.Printf("processing source: %s", uri)
			rc, err := fetchSource(uri)
			if err != nil {
				Log.Error.Printf("failed to fetch %s: %v", uri, err)
				continue
			}
			events := processSource(rc, now, conf)
			allEvents = append(allEvents, events...)
			rc.Close()
		}
	} else {
		// Stdin fallback
		Log.Info.Println("reading from stdin")
		allEvents = processSource(os.Stdin, now, conf)
	}

	// Sort events by Start date in DESCENDING order
	sort.Slice(allEvents, func(i, j int) bool {
		ti := rawToStructDate(allEvents[i].Start)
		tj := rawToStructDate(allEvents[j].Start)
		return ti.After(tj)
	})

	if conf.Limit > 0 && len(allEvents) > conf.Limit {
		allEvents = allEvents[:conf.Limit]
	}

	// Output minified JSON
	jsonBytes, err := json.Marshal(allEvents)
	if err != nil {
		Log.Error.Fatalf("failed to marshal events: %v", err)
	}

	if fileSet {
		target := resolveOutputPath(filePath)
		Log.Info.Printf("writing to %s", target)
		if err := safeWrite(target, jsonBytes); err != nil {
			Log.Error.Fatalf("failed to write output: %v", err)
		}
	} else {
		fmt.Println(string(jsonBytes))
	}
}
