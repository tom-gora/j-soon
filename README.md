# JSON-from-iCal

**JSON-from-iCal just produces JSON with your upcoming calendar events. UNIX philosphy. Period.**  

This project is a utility that extracts upcoming events from iCal calendar files (.ics) and outputs into JSON format. It is made to be scriptable to fit into your pipeline, indifferent to how the resulting data is consumed. It can take stdin string and both local and remote iCal files as declared in basic config file. It outputs json either to stdout or to a default / selected location .json file.  

## Core Philosophy

The Go-based core is a pure data producer. It's core uses excellent `arran4/golang-ical`.  

It does basic filtering, arithmetic and string formatting (to my liking, at this point not configurable at the MVP level unles you are willing to modify and run `go build` yourself), then outputs a minified JSON array of events from a given range of upcoming days. Whether this data is piped into your exquisitly styled Linux notification daemon popups, a web dashboard or a Discord webhook does not matter. Not one (go) iota. 😉  

### Personal use case for making this: System Notifications

While the binary is general-purpose, this repository includes an auxiliary pipeline (a small shell script using `jq` to process output and pipe data into `notify-send`) for a specific desktop use case: **System Notifications**. (Note: Project is personal and any scripts are reference impl)  

## Origin and attributions

This project was a fork of `jceaser/readical`, but has undergone a lot of changes to tidy it up and change its core functionality to my needs.  

- **JSON-Only**: Completely stripped all markup and formatting logic from the "inspiration project".  
- **Upcoming Time Window**: Replaced complex time-ranging with a single, intuitive `--upcoming-days` lookahead.  
- **Idiomatic Go**: Refactored due to barely any adherance to go standards in the base project half items Exported, halp private, 3 different text cases in one func signature and things like that...  

## Key Features

- **Standardized Schema**: Enforced thanks to go's first-class-citizen (almost JS level) json handling.  
- **Flexible Ingestion**: Supports local `.ics` files, remote `http/https` URLs, and direct `stdin` piping.  
- **Noise Reduction**: Automatically filters out iCal clutter (e.g., `X-APPLE-` metadata). Gets you readable event data in no-time.  

## Build and Testing

### Build

```bash
make build
```  

The binary is written to `bin/cal-event-notifier`.  

### Testing

This project features quite a wide test suite covering both go internals and external interactions (scripts, piping, processing config etc).  
_**Note**: Unlike the code adaptation and writing, the tests were created by an AI agent under the strong and specific guidance of my watchful eye. Don't we know it, nobody likes writing tests. But designing and reviewing them may just be ok 😉_  

- **Unit Tests**: Verifies internal date parsing and time window logic (programmatically, not based on hardcoded data).  
- **Integrity Tests**: Validates JSON structure, limits, and CLI flag behavior.  

```bash
make test_all
```  

---  

_A focused tool for data-driven calendar workflows._
