<p aria-label="j-soon" disabled align="center">  
  <img disabled width="90%" src="./_assets/header.svg" alt="j-soon"><br>  
</p>  

![fun-with-skeumorphics](./_assets/image.png)  
<br><br>  

**J-SOON just produces JSON with your upcoming calendar events. Nothing more. UNIX philosphy. Period.**  

This project is a utility that extracts upcoming events from iCal calendar files (.ics) and outputs into JSON format. It is made to be scriptable to fit into your pipeline, indifferent to how the resulting data is consumed. It can take stdin string and both local and remote iCal files as declared in basic config file. It outputs json either to stdout or to a default / selected location .json file.  

## Core Philosophy

The Go-based binary at the core is a pure data producer. It uses `arran4/golang-ical` library for calendar parsing.  

It does basic filtering, arithmetic and string formatting (to my liking, at this point not configurable at the MVP level unless you are willing to modify and run `go build` yourself), then outputs a minified JSON array of events from a given range of upcoming days. Whether this data is piped into your exquisitely styled Linux notification daemon popups, a web dashboard or a Discord webhook does not matter. Not one (go) iota. 😉  

**Personal use case for making this: System Notifications**  

https://github.com/user-attachments/_assets/09cd5410-273d-4fce-bcae-103caa3e3eba  

While the binary is general-purpose, this repository includes an auxiliary pipeline (a small shell script using `jq` to process output and pipe data into `notify-send`) for a specific desktop use case: **System Notifications**. (Note: Project is personal and any scripts are just a reference implementation)  

## Key Features

- **Standardized Schema**: Enforced thanks to go's first-class-citizen (almost JS level) json handling.  
- **Flexible Ingestion**: Supports local `.ics` files, remote `http/https` URLs, and direct `stdin` piping.  

## Usage

`jsoon` automatically detects its input mode based on whether data is being piped to it.  

### 1. Stdin Mode is Primary Mode

If data is piped via **Stdin**, `jsoon` processes it and ignores any calendars defined in the configuration file. Good for scripting.  

```bash
cat calendar.ics | ./bin/jsoon
```  

### 2. Config Mode

If no data is detected on **Stdin**, `jsoon` looks for calendars in its configuration file.  

Pass a JSON configuration file containing paths to local `.ics` files or remote URLs:  

```json
{
  "calendars": [
    "~/path/to/local/calendar.ics",
    "https://remote-resource.com/fetch-me.ics"
  ],
  "upcoming_days": 7,
  "events_limit": 0
}
```  

```bash
./bin/jsoon --config config.json
```  

**Persistence**: Default config path is `$XDG_CONFIG_HOME/jsoon/config.json` (falls back to `~/.config/jsoon/config.json`). If the file does not exist, `jsoon` will create it with default values on the first run.  

#### Default `config.json` Structure

When bootstrapped, your configuration will look like this:  

```json
{
  "calendars": [],
  "upcoming_days": 7,
  "events_limit": 0,
  "output_file": "",
  "date_template": "YYYY MMM DD",
  "offset_markers": {}
}
```  

| Key              | Description                                                                                     |  
| :--------------- | :---------------------------------------------------------------------------------------------- |  
| `calendars`      | An array of strings (URLs or local file paths) to process.                                      |  
| `upcoming_days`  | Number of days from today to look ahead.                                                        |  
| `events_limit`   | Max number of total events to return (0 = unlimited).                                           |  
| `output_file`    | Path to write output JSON. If empty or `"stdout"`, prints to terminal. Supports `~/` expansion. |  
| `date_template`  | Template the format of your choice of how to show dates.                                        |  
| `offset_markers` | A map of day offsets to string suffixes to append to the date.                                  |  

#### Offset Markers (Dynamic Suffixes)

You can configure special suffixes to be appended to the formatted date string based on how many days away the event is. This is useful for highlighting events that are "TODAY" or "TOMORROW".  

```json
"offset_markers": {
  "0": " - TODAY !!",
  "1": " - TOMORROW !",
  "7": " - IN A WEEK"
}
```  

- Keys must be strings representing the integer offset from today (0 = today, 1 = tomorrow, etc.).  
- The corresponding value will be appended to the date string if the event matches that day.  
- This feature is **only** configurable via the config file.  

#### Customizing Date Templates

`jsoon` uses the [fmtdate](https://gitlab.com/metakeule/fmtdate) library for date formatting. You can change how your dates look by modifying the `date_template` in your config or using the `-t` flag.  

**Common Placeholders:**  

| Placeholder | Meaning             | Example    |  
| :---------- | :------------------ | :--------- |  
| `YYYY`      | 4-digit year        | `2026`     |  
| `MM`        | Month (numeric)     | `02`       |  
| `MMM`       | Month (short name)  | `Feb`      |  
| `MMMM`      | Month (full name)   | `February` |  
| `DD`        | Day of month        | `01`       |  
| `D`         | Day of month (lean) | `1`        |  
| `DDD`       | Day of week (short) | `Sun`      |  
| `DDDD`      | Day of week (full)  | `Sunday`   |  
| `hh`        | Hour (24h)          | `14`       |  
| `mm`        | Minutes             | `05`       |  

_Note: Any characters not matching placeholders (like `[ ]`, `/`, or `-`) are preserved as-is._  

### CLI Flags

| Flag              | Short | Default  | Description                             |  
| :---------------- | :---- | :------- | :-------------------------------------- |  
| `--upcoming-days` | `-u`  | `7`      | Days ahead to look for events           |  
| `--limit`         | `-l`  | `0`      | Max number of events (0 = unlimited)    |  
| `--output-file`   | `-f`  | `stdout` | Output file path (defaults to terminal) |  
| `--config`        | `-c`  | `""`     | Path to custom config.json              |  
| `--template`      | `-t`  | `""`     | Template string for output dates        |  
| `--verbose`       | `-v`  | `false`  | Enable detailed logging                 |  
| `--version`       | `-V`  | `false`  | Show version information                |  

### Example

```bash
./bin/jsoon -u 2 -l 1 -c ./test_data/test_config.json | jq
```  

## Build and Testing

### Build

```bash
make build
```  

The binary is written to `bin/jsoon`. Place it in your `$PATH` according to your preference.  

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
