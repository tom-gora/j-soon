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
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/tom-gora/j-soon/internal/config"
	fu "github.com/tom-gora/j-soon/internal/fileutil"
	"github.com/tom-gora/j-soon/internal/ical"
	l "github.com/tom-gora/j-soon/internal/logger"
)

func main() {
	// init with defaults
	eCtx, uris := config.InitCtx()
	allEvents := []ical.CalendarEvent{}
	now := time.Now()

	if eCtx.IsStdin {
		// Stdin fallback
		l.Log.Info.Println("reading from stdin")
		allEvents = ical.ProcessSourceToStruct(os.Stdin, now, eCtx.DateTemplate, eCtx.UpcomingDays, eCtx.OffsetMarkers)
	} else {
		for _, uri := range uris {
			l.Log.Info.Printf("processing source: %s", uri)
			rc, err := ical.FetchSource(uri)
			if err != nil {
				l.Log.Error.Printf("failed to fetch %s: %v", uri, err)
				continue
			}
			events := ical.ProcessSourceToStruct(rc, now, eCtx.DateTemplate, eCtx.UpcomingDays, eCtx.OffsetMarkers)
			allEvents = append(allEvents, events...)
			rc.Close()
		}
	}

	// Sort events by Start date in DESCENDING order
	sort.Slice(allEvents, func(i, j int) bool {
		ti := ical.StrToStructDate(allEvents[i].Start)
		tj := ical.StrToStructDate(allEvents[j].Start)
		return ti.After(tj)
	})

	if eCtx.Limit > 0 && len(allEvents) > eCtx.Limit {
		allEvents = allEvents[:eCtx.Limit]
	}

	// Output minified JSON
	jsonBytes, err := json.Marshal(allEvents)
	if err != nil {
		l.Log.Error.Fatalf("failed to marshal events: %v", err)
	}

	// Determine output target
	target := fu.GetOutputPath(eCtx.OutputFile)

	if target != "" && target != "stdout" {
		l.Log.Info.Printf("writing to %s", target)
		if err := fu.SafeWrite(target, jsonBytes); err != nil {
			l.Log.Error.Fatalf("failed to write output: %v", err)
		}
	} else {
		fmt.Println(string(jsonBytes))
	}
}
