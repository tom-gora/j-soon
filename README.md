<p align="center">  
  <img width="80%" src="./assets/header.svg" alt="header">  
</p>  
<br>  
<br>  

![fun-with-skeumorphics](./assets/image.png)  
<br><br>  

**JSON-from-iCal just produces JSON with your upcoming calendar events. Nothing more. UNIX philosphy. Period.**  

This project is a utility that extracts upcoming events from iCal calendar files (.ics) and outputs into JSON format. It is made to be scriptable to fit into your pipeline, indifferent to how the resulting data is consumed. It can take stdin string and both local and remote iCal files as declared in basic config file. It outputs json either to stdout or to a default / selected location .json file.  

## Core Philosophy

The Go-based binary at the core is a pure data producer. It uses `arran4/golang-ical` library for calendar parsing.  

It does basic filtering, arithmetic and string formatting (to my liking, at this point not configurable at the MVP level unless you are willing to modify and run `go build` yourself), then outputs a minified JSON array of events from a given range of upcoming days. Whether this data is piped into your exquisitely styled Linux notification daemon popups, a web dashboard or a Discord webhook does not matter. Not one (go) iota. 😉  

**Personal use case for making this: System Notifications**  

https://github.com/user-attachments/assets/09cd5410-273d-4fce-bcae-103caa3e3eba  

While the binary is general-purpose, this repository includes an auxiliary pipeline (a small shell script using `jq` to process output and pipe data into `notify-send`) for a specific desktop use case: **System Notifications**. (Note: Project is personal and any scripts are reference impl)  

## Key Features

- **Standardized Schema**: Enforced thanks to go's first-class-citizen (almost JS level) json handling.  
- **Flexible Ingestion**: Supports local `.ics` files, remote `http/https` URLs, and direct `stdin` piping.  
- **Noise Reduction**: Automatically filters out iCal clutter (e.g., `X-APPLE-` metadata). Gets you readable event data in no-time.  

## Usage

`jfi` operates in two primary modes: **Stdin** and **Config**.  

- **Input Default**: If no `--config` is provided, the tool reads from `stdin`.  
- **Output Default**: If no `--file` is provided, the tool writes to `stdout`.  

### 1. Stdin Mode

Pipe iCal data directly into the tool when scripting.  

```bash
cat calendar.ics | ./bin/jfi
```  

### 2. Config Mode

Pass a simple configuration file containing paths to local `.ics` files or remote URLs (one per line):  

```
# comments allowed
~/path/to/local/calendar.ics

https://remote-resource.com/fetch-me.ics
```  

```bash
./bin/jfi --config calendars.conf
```  

### CLI Flags

| Flag              | Short | Default  | Description                           |  
| :---------------- | :---- | :------- | :------------------------------------ |  
| `--upcoming-days` | `-u`  | `7`      | Days ahead to look for events         |  
| `--limit`         | `-l`  | `0`      | Max number of events (0 = unlimited)  |  
| `--file`          | `-f`  | `stdout` | Output file path (see Priority Logic) |  
| `--config`        | `-c`  | `""`     | Path to calendars configuration file  |  
| `--verbose`       | `-v`  | `false`  | Enable detailed logging (! WIP !)     |  
| `--version`       | `-V`  | `false`  | Show version information              |  

### Example

Without the `-f`/`--file` flag, the tool outputs JSON to `stdout`:  

```bash
./bin/jfi -u 2 -l 1 -c ./test_data/test_calendars.conf | jq
```  

Output:  

```json
[
  {
    "UID": "bday-charlie",
    "Start": "20260124",
    "HumanStart": "[ SAT ] 24 Jan 2026 @ 00:00",
    "End": "20260125",
    "HumanEnd": "[ SUN ] 25 Jan 2026 @ 00:00",
    "ActualEnd": "20260125",
    "Summary": "Charlie's Birthday",
    "Location": "",
    "Description": "",
    "Hours": 24,
    "SubDay": false,
    "Day": true,
    "MultiDay": false,
    "Ongoing": false
  }
]
```  

### Output Priority Logic

When the `--file` flag is provided without a value (e.g., `./jfi -f`), the tool determines the output path in this order:  

1. `$XDG_CACHE_HOME/event-notifications/out.json`  
2. `$HOME/.cache/event-notifications/out.json`  
3. `./out/out.json`  

## Build and Testing

### Build

```bash
make build
```  

The binary is written to `bin/jfi`. Place it in your `$PATH` according to your preference.  

### Testing

This project features quite a wide test suite covering both go internals and external interactions (scripts, piping, processing config etc).  
_**Note**: Unlike the code adaptation and writing, the tests were created by an AI agent under the strong and specific guidance of my watchful eye. Don't we know it, nobody likes writing tests. But designing and reviewing them may just be ok_ 😉  

- **Unit Tests**: Verifies internal date parsing and time window logic (programmatically, not based on hardcoded data).  
- **Integrity Tests**: Validates JSON structure, limits, and CLI flag behavior.  

```bash
make test_all
```  

## Attribution

This project was initially a fork of [jceaser/readical](https://github.com/jceaser/readical), but has undergone a lot of changes to tidy it up and change its core functionality to my needs to the point it does not relate much anymore.  

## License

This project is published under the **MIT License**. See the [LICENSE](./LICENSE) file for more details.  

---  

_A focused tool for data-driven calendar workflows._
