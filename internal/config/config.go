// Package config
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	fu "github.com/tom-gora/j-soon/internal/fileutil"
	l "github.com/tom-gora/j-soon/internal/logger"
)

// Defaults

const (
	AppBy      = "J-SOON by github.com/tom-gora"
	AppVersion = "1.0.0"
)

type ExitCode int

const (
	ExitNorm ExitCode = iota
	ExitVer
)

func (ec ExitCode) Int() int {
	return int(ec)
}

const (
	DefaultUpcomingDays = 7
	DefaultLimit        = 0
	DefaultOutputFile   = ""
	DefaultDateTemplate = "YYYY MMM DD"
	DefaultVerbose      = false
)

type ConfigModel struct {
	Calendars     []string          `json:"calendars"`
	UpcomingDays  int               `json:"upcoming_days"`
	Limit         int               `json:"events_limit"`
	OutputFile    string            `json:"output_file"`
	DateTemplate  string            `json:"date_template"`
	OffsetMarkers map[string]string `json:"offset_markers"`
}

type ExecutionCtx struct {
	IsStdin       bool
	Calendars     []string
	UpcomingDays  int
	Limit         int
	OutputFile    string
	ConfigPath    string
	DateTemplate  string
	Verbose       bool
	OffsetMarkers map[string]string
}

type FlagCtx struct {
	UpcomingDays    int
	Limit           int
	OutputFile      string
	ConfigPath      string
	DateTemplate    string
	Verbose         bool
	IsOutputFileSet bool
	ShowVersion     bool
	SpecifiedFlags  map[string]bool
}

func getDefaultConfigValues() ConfigModel {
	return ConfigModel{
		Calendars:     []string{},
		UpcomingDays:  DefaultUpcomingDays,
		Limit:         DefaultLimit,
		OutputFile:    DefaultOutputFile,
		DateTemplate:  DefaultDateTemplate,
		OffsetMarkers: map[string]string{},
	}
}

func usage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, "\n%s\n", color.New(color.FgCyan, color.Bold).Sprint(AppBy))
	fmt.Fprintf(out, "Version: %s\n\n", color.New(color.FgYellow).Sprint(AppVersion))
	fmt.Fprintf(out, "%s\n", color.New(color.FgHiWhite, color.Underline).Sprint("Usage:"))
	fmt.Fprintf(out, "  jsoon [flags]\n\n")

	headerFmt := color.New(color.FgHiGreen, color.Bold).SprintfFunc()
	columnFmt := color.New(color.FgHiWhite).SprintfFunc()

	tbl := table.New("Flag", "Short", "Default", "Description")
	tbl.WithWriter(out)
	tbl.WithHeaderFormatter(headerFmt)
	tbl.WithFirstColumnFormatter(columnFmt)

	flags := []struct {
		long, short, def, desc string
	}{
		{"config", "c", "", "Path to custom config.json"},
		{"upcoming-days", "u", "7", "Number of upcoming days to include (7=default)"},
		{"limit", "l", "0", "Max number of events to process (0=unlimited)"},
		{"output-file", "f", "stdout", "Output file (literal \"stdout\" or empty for priority logic)"},
	}

	for _, f := range flags {
		tbl.AddRow("--"+f.long, "-"+f.short, f.def, f.desc)
	}

	tbl.Print()
	fmt.Fprintln(out)
}

func setDefaultConfigPath() (string, error) {
	var configPath string
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig != "" {
		configPath = filepath.Join(xdgConfig, "jsoon", "config.json")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configPath = filepath.Join(home, ".config", "jsoon", "config.json")
	}
	dir := filepath.Dir(configPath)
	_, err := os.Stat(configPath)
	if err != nil {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return "", err
		}
		contentStruct := getDefaultConfigValues()
		contentBytes, _ := json.MarshalIndent(contentStruct, "", "  ")
		err := os.WriteFile(configPath, contentBytes, 0o644)
		if err != nil {
			return "", err
		}
	}
	return configPath, nil
}

func getFileConfig(p string) (ConfigModel, error) {
	config := ConfigModel{
		Calendars:     []string{},
		UpcomingDays:  DefaultUpcomingDays,
		Limit:         DefaultLimit,
		OutputFile:    DefaultOutputFile,
		DateTemplate:  DefaultDateTemplate,
		OffsetMarkers: map[string]string{},
	}

	JSONBytes, err := os.ReadFile(fu.PathExpandTilde(p))
	if err != nil {
		return ConfigModel{}, err
	}

	err = json.Unmarshal(JSONBytes, &config)
	if err != nil {
		return ConfigModel{}, err
	}
	return config, nil
}

func initFlags(fCtx *FlagCtx) {
	flag.Usage = usage
	flag.StringVar(&fCtx.ConfigPath, "config", "", "")
	flag.StringVar(&fCtx.ConfigPath, "c", "", "")
	flag.IntVar(&fCtx.UpcomingDays, "upcoming-days", 7, "")
	flag.IntVar(&fCtx.UpcomingDays, "u", 7, "")
	flag.IntVar(&fCtx.Limit, "limit", 0, "")
	flag.IntVar(&fCtx.Limit, "l", 0, "")
	flag.StringVar(&fCtx.DateTemplate, "template", "", "")
	flag.StringVar(&fCtx.DateTemplate, "t", "", "")
	flag.BoolVar(&fCtx.Verbose, "verbose", false, "")
	flag.BoolVar(&fCtx.Verbose, "v", false, "")
	flag.BoolVar(&fCtx.ShowVersion, "version", false, "")
	flag.BoolVar(&fCtx.ShowVersion, "V", false, "")

	fCtx.IsOutputFileSet = false
	// Custom flag parsing for -f to detect if it's set at all
	flag.Func("output-file", "Output file", func(s string) error {
		fCtx.OutputFile = s
		fCtx.IsOutputFileSet = true
		return nil
	})
	flag.Func("f", "Output file", func(s string) error {
		fCtx.OutputFile = s
		fCtx.IsOutputFileSet = true
		return nil
	})
	flag.Func("file", "Output file", func(s string) error {
		fCtx.OutputFile = s
		fCtx.IsOutputFileSet = true
		return nil
	})

	flag.Parse()

	fCtx.SpecifiedFlags = make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		fCtx.SpecifiedFlags[f.Name] = true
	})
}

func InitCtx() (ExecutionCtx, []string) {
	fCtx := FlagCtx{}
	initFlags(&fCtx)

	confFile := fCtx.ConfigPath
	if confFile == "" {
		var err error
		confFile, err = setDefaultConfigPath()
		if err != nil {
			l.Log.Error.Fatalf("failed to set up default config environment: %v. Please ensure $HOME or $XDG_CONFIG_HOME is correctly set and writable.", err)
		}
	}

	// 1. Set default
	CONTEXT := ExecutionCtx{
		IsStdin:       false,
		Calendars:     []string{},
		UpcomingDays:  DefaultUpcomingDays,
		Limit:         DefaultLimit,
		OutputFile:    DefaultOutputFile,
		ConfigPath:    confFile,
		DateTemplate:  DefaultDateTemplate,
		Verbose:       DefaultVerbose,
		OffsetMarkers: map[string]string{},
	}

	// 2. check the config file and override if new config values
	fileConfigCtx, err := getFileConfig(CONTEXT.ConfigPath)
	if err != nil {
		lerr := fmt.Sprintf("failed parsing config file: %s", CONTEXT.ConfigPath)
		l.Log.Error.Fatalf("%s: %v", lerr, err)
		usage()
		os.Exit(1)
	}
	uris := fileConfigCtx.Calendars
	// Merge file values into CONTEXT
	if fileConfigCtx.UpcomingDays > 0 {
		CONTEXT.UpcomingDays = fileConfigCtx.UpcomingDays
	}
	if fileConfigCtx.Limit > 0 {
		CONTEXT.Limit = fileConfigCtx.Limit
	}
	if fileConfigCtx.OutputFile != "" {
		CONTEXT.OutputFile = fileConfigCtx.OutputFile
	}
	if fileConfigCtx.DateTemplate != "" {
		CONTEXT.DateTemplate = fileConfigCtx.DateTemplate
	}
	if fileConfigCtx.OffsetMarkers != nil {
		CONTEXT.OffsetMarkers = fileConfigCtx.OffsetMarkers
	}

	// 3. carry on and finish with overrides from explicit flags
	if fCtx.SpecifiedFlags["u"] || fCtx.SpecifiedFlags["upcoming-days"] {
		CONTEXT.UpcomingDays = fCtx.UpcomingDays
	}
	if fCtx.SpecifiedFlags["l"] || fCtx.SpecifiedFlags["limit"] {
		CONTEXT.Limit = fCtx.Limit
	}
	if fCtx.IsOutputFileSet {
		CONTEXT.OutputFile = fCtx.OutputFile
	}
	if fCtx.SpecifiedFlags["t"] || fCtx.SpecifiedFlags["template"] {
		CONTEXT.DateTemplate = fCtx.DateTemplate
	}
	if fCtx.SpecifiedFlags["v"] || fCtx.SpecifiedFlags["verbose"] {
		CONTEXT.Verbose = fCtx.Verbose
	}

	stat, _ := os.Stdin.Stat()
	isStdin := stat.Mode()&os.ModeCharDevice == 0
	CONTEXT.IsStdin = isStdin

	if isStdin {
		// If data is piped, we ignore the calendars from config
		uris = []string{}
	}

	if fCtx.ShowVersion {
		fmt.Printf("%s\nVersion %s\n", AppBy, AppVersion)
		os.Exit(ExitNorm.Int())
	}

	if CONTEXT.Verbose {
		l.Log.EnableInfo()
		l.Log.EnableDebug()
	}

	if len(uris) < 1 && !isStdin {
		l.Log.Error.Fatal("input must be provided. Either configure your calendars in the config file, or feed the content of a calendar file via stdin")
	}

	return CONTEXT, uris
}
