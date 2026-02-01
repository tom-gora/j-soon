// Package config
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/tom-gora/JSON-from-iCal/internal/common"
	l "github.com/tom-gora/JSON-from-iCal/internal/logger"
)

// Defaults
const (
	DefaultUpcomingDays = 7
	DefaultLimit        = 0
	DefaultOutputFile   = "./out/jfi_out.json"
	DefaultDateTemplate = "YYYY MMM DD"
	DefaultVerbose      = false
)

type ConfigModel struct {
	Calendars    []string `json:"calendars"`
	UpcomingDays int      `json:"upcoming_days"`
	Limit        int      `json:"events_limit"`
	OutputFile   string   `json:"output_file"`
	DateTemplate string   `json:"date_template"`
}

type ExecutionCtx struct {
	IsStdin      bool
	Calendars    []string
	UpcomingDays int
	Limit        int
	OutputFile   string
	ConfigPath   string
	DateTemplate string
	Verbose      bool
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
		Calendars:    []string{},
		UpcomingDays: DefaultUpcomingDays,
		Limit:        DefaultLimit,
		OutputFile:   DefaultOutputFile,
		DateTemplate: DefaultDateTemplate,
	}
}

func usage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, "\n%s\n", color.New(color.FgCyan, color.Bold).Sprint(common.AppBy))
	fmt.Fprintf(out, "Version: %s\n\n", color.New(color.FgYellow).Sprint(common.AppVersion))
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
		{"config", "c", "", "Path to custom config.json"},
		{"upcoming-days", "u", "7", "Number of upcoming days to include (7=default)"},
		{"limit", "l", "0", "Max number of events to process (0=unlimited)"},
		{"output-file", "o", "stdout", "Output file (empty for priority logic)"},
		{"template", "t", "stdout", "Template string to format output date"},
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

func setDefaultConfigPath() (string, error) {
	var configPath string
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig != "" {
		configPath = filepath.Join(xdgConfig, "jfi", "config.json")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configPath = filepath.Join(home, ".config", "jfi", "config.json")
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

func isExplicitConfigPathPassed() (string, error) {
	value := ""
	args := os.Args
	for i := 1; i < len(args); i++ {
		// ignore other args
		if args[i] != "-c" && args[i] != "--config" && args[i] != "-config" {
			continue
		}
		// if we match try to grab arg to flag if exists as long as it is not another flag
		if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
			value = strings.TrimSpace(args[i+1])
		}
		// if nothing followed and next arg was another flag
		// or -c was last and used w/o arg
		if (i < len(args)-1 && strings.HasPrefix(args[i+1], "-")) || (i == len(args)-1 && strings.TrimSpace(value) == "") {
			return value, errors.New("flag -c / --config requires passing a path to a valid config file")
		}
	}
	return value, nil
}

func getFileConfig(p string) (ConfigModel, error) {
	config := ConfigModel{
		Calendars:    []string{},
		UpcomingDays: DefaultUpcomingDays,
		Limit:        DefaultLimit,
		OutputFile:   DefaultOutputFile,
		DateTemplate: DefaultDateTemplate,
	}

	JSONBytes, err := os.ReadFile(common.PathExpandTilde(p))
	if err != nil {
		return ConfigModel{}, err
	}

	err = json.Unmarshal(JSONBytes, &config)
	if err != nil {
		return ConfigModel{}, err
	}
	return config, nil
}

func SetCtxFromConfig(ec *ExecutionCtx, fc FlagCtx) ([]string, error) {
	if fc.ConfigPath == "" {
		return []string{}, nil
	}

	ec.ConfigPath = fc.ConfigPath

	JSONBytes, err := os.ReadFile(common.PathExpandTilde(fc.ConfigPath))
	if err != nil {
		return []string{}, err
	}

	var config ConfigModel
	err = json.Unmarshal(JSONBytes, &config)
	if err != nil {
		return []string{}, err
	}

	if config.UpcomingDays > 0 {
		ec.UpcomingDays = config.UpcomingDays
	}
	if config.Limit > 0 {
		ec.Limit = config.Limit
	}
	if config.OutputFile != "" {
		ec.OutputFile = config.OutputFile
	}
	if config.DateTemplate != "" {
		ec.DateTemplate = config.DateTemplate
	}
	return config.Calendars, nil
}

func initFlags(fCtx *FlagCtx) {
	flag.Usage = usage
	flag.StringVar(&fCtx.ConfigPath, "config", "", "")
	flag.StringVar(&fCtx.ConfigPath, "c", "", "")
	flag.IntVar(&fCtx.UpcomingDays, "upcoming-days", 7, "")
	flag.IntVar(&fCtx.UpcomingDays, "u", 7, "")
	flag.IntVar(&fCtx.Limit, "limit", 0, "")
	flag.IntVar(&fCtx.Limit, "l", 0, "")
	flag.StringVar(&fCtx.OutputFile, "output-file", "", "")
	flag.StringVar(&fCtx.OutputFile, "o", "", "")
	flag.StringVar(&fCtx.DateTemplate, "template", "", "")
	flag.StringVar(&fCtx.DateTemplate, "t", "", "")
	flag.BoolVar(&fCtx.Verbose, "verbose", false, "")
	flag.BoolVar(&fCtx.Verbose, "v", false, "")
	flag.BoolVar(&fCtx.ShowVersion, "version", false, "")
	flag.BoolVar(&fCtx.ShowVersion, "V", false, "")

	fCtx.IsOutputFileSet = false

	flag.Parse()
}

func setFlagCtx(fCtx *FlagCtx) {
	// Custom flag parsing for -f to detect if it's set at all
	flag.Func("file", "Output file", func(s string) error {
		fCtx.OutputFile = s
		fCtx.IsOutputFileSet = true
		return nil
	})

	flag.Func("f", "Output file", func(s string) error {
		fCtx.OutputFile = s
		fCtx.IsOutputFileSet = true
		return nil
	})
	fCtx.SpecifiedFlags = make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		fCtx.SpecifiedFlags[f.Name] = true
	})
}

func InitCtx() ExecutionCtx {
	confFile, err := isExplicitConfigPathPassed()
	if err != nil {
		l.Log.Error.Printf("wrong use of -c | --config flag: %v\n", err)
		usage()
	}
	if confFile != "" && !common.IsValidFile(confFile) {
		l.Log.Error.Fatalf("passed config appears to not be a valid file: %v", err)
		os.Exit(1)
	}
	if confFile == "" {
		var err error
		confFile, err = setDefaultConfigPath()
		if err != nil {
			l.Log.Error.Fatalf("failed to set up default config environment: %v. Please ensure $HOME or $XDG_CONFIG_HOME is correctly set and writable.", err)
		}
	}
	// 1. Set default
	CONTEXT := ExecutionCtx{
		IsStdin:      false,
		Calendars:    []string{},
		UpcomingDays: DefaultUpcomingDays,
		Limit:        DefaultLimit,
		OutputFile:   DefaultOutputFile,
		ConfigPath:   confFile,
		DateTemplate: DefaultDateTemplate,
		Verbose:      DefaultVerbose,
	}

	// 2. check the config file and override if new config files
	fileConfigCtx, err := getFileConfig(CONTEXT.ConfigPath)
	if err != nil {
		lerr := fmt.Sprintf("failed parsing file:%s", CONTEXT.ConfigPath)
		l.Log.Error.Fatalf(lerr, err)
		usage()
		os.Exit(1)
	}

	stat, _ := os.Stdin.Stat()
	isStdin := stat.Mode()&os.ModeCharDevice == 0

	if len(fileConfigCtx.Calendars) < 1 && !isStdin {
		l.Log.Error.Fatal("input must be provided. Either configure your calendars in the config file, or feed the content of a calendar file via stdin", err)
	}

	//
	// -- carry on and finish with overrides from explicit flags
	fCtx := FlagCtx{}
	initFlags(&fCtx)
	setFlagCtx(&fCtx)

	if fCtx.SpecifiedFlags["u"] || fCtx.SpecifiedFlags["upcoming-days"] {
		CONTEXT.UpcomingDays = fCtx.UpcomingDays
	}
	if fCtx.SpecifiedFlags["l"] || fCtx.SpecifiedFlags["limit"] {
		CONTEXT.Limit = fCtx.Limit
	}
	if fCtx.SpecifiedFlags["o"] || fCtx.SpecifiedFlags["output-file"] {
		CONTEXT.OutputFile = fCtx.OutputFile
	}
	if fCtx.SpecifiedFlags["t"] || fCtx.SpecifiedFlags["template"] {
		CONTEXT.DateTemplate = fCtx.DateTemplate
	}
	CONTEXT.IsStdin = isStdin
	jsonBytes, _ := json.MarshalIndent(CONTEXT, "", "  ")
	fmt.Print(string(jsonBytes))
	return CONTEXT
}
