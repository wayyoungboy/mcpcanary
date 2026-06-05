package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/wayyoungboy/mcpcanary/internal/analyzer"
	"github.com/wayyoungboy/mcpcanary/internal/config"
	"github.com/wayyoungboy/mcpcanary/internal/lockfile"
	"github.com/wayyoungboy/mcpcanary/internal/model"
	"github.com/wayyoungboy/mcpcanary/internal/report"
	"github.com/wayyoungboy/mcpcanary/internal/trust"
)

const Version = "0.1.2"

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printHelp(stdout)
		return 0
	}
	switch args[0] {
	case "scan":
		return runScan(args[1:], stdout, stderr)
	case "inspect":
		return runInspect(args[1:], stdout, stderr)
	case "lock":
		return runLock(args[1:], stdout, stderr)
	case "diff":
		return runDiff(args[1:], stdout, stderr)
	case "version":
		fmt.Fprintf(stdout, "MCPCanary %s\n", Version)
		return 0
	case "help", "-h", "--help":
		printHelp(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		printHelp(stderr)
		return 2
	}
}

func runScan(args []string, stdout, stderr io.Writer) int {
	parsed, err := parseOptions(args, map[string]string{"format": "text", "fail-on": "", "trust-store": defaultTrustStorePath()})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	paths := parsed.Positionals
	if len(paths) == 0 {
		paths = config.Discover(homeDir())
	}
	if len(paths) == 0 {
		fmt.Fprintln(stderr, "no MCP config paths provided or discovered")
		return 2
	}
	scan, err := scanPaths(paths, parsed.Values["trust-store"])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	output, err := report.Render(scan, report.Format(parsed.Values["format"]))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	fmt.Fprintln(stdout, output)
	if exceeds(scan.Findings, parsed.Values["fail-on"]) {
		return 1
	}
	return 0
}

func runInspect(args []string, stdout, stderr io.Writer) int {
	cfg, err := loadSingle(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	data, err := json.MarshalIndent(cfg.Servers, "", "  ")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintln(stdout, string(data))
	return 0
}

func runLock(args []string, stdout, stderr io.Writer) int {
	parsed, err := parseOptions(args, map[string]string{"lockfile": "mcpcanary.lock"})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	cfg, err := loadSingle(parsed.Positionals)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	scan := analyzer.Analyze(cfg.Servers, analyzer.Options{})
	lock := lockfile.FromServers(cfg.Servers, scan.Score)
	if err := lockfile.Write(parsed.Values["lockfile"], lock); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "wrote %s with %d server(s)\n", parsed.Values["lockfile"], len(lock.Servers))
	return 0
}

func runDiff(args []string, stdout, stderr io.Writer) int {
	parsed, err := parseOptions(args, map[string]string{"lockfile": "mcpcanary.lock"})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	cfg, err := loadSingle(parsed.Positionals)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	lock, err := lockfile.Read(parsed.Values["lockfile"])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	diff := lockfile.Compare(lock, cfg.Servers, lockfile.CompareOptions{DriftThreshold: 0.18})
	if len(diff.Changes) == 0 {
		fmt.Fprintln(stdout, "no drift")
		return 0
	}
	for _, change := range diff.Changes {
		if change.Distance > 0 {
			fmt.Fprintf(stdout, "%s: %s (distance %.3f)\n", change.Kind, change.ServerName, change.Distance)
			continue
		}
		fmt.Fprintf(stdout, "%s: %s\n", change.Kind, change.ServerName)
	}
	return 1
}

func scanPaths(paths []string, trustStorePath string) (model.Report, error) {
	var servers []model.Server
	var firstPath string
	for _, path := range paths {
		cfg, err := config.LoadFile(path)
		if err != nil {
			return model.Report{}, err
		}
		if firstPath == "" {
			firstPath = cfg.Path
		}
		servers = append(servers, cfg.Servers...)
	}
	store, err := trust.Open(trustStorePath)
	if err != nil {
		return model.Report{}, err
	}
	scan := analyzer.Analyze(servers, analyzer.Options{SemanticMatcher: store})
	scan.ConfigPath = firstPath
	return scan, nil
}

func loadSingle(args []string) (model.Config, error) {
	if len(args) != 1 {
		return model.Config{}, fmt.Errorf("expected exactly one config path")
	}
	return config.LoadFile(args[0])
}

func exceeds(findings []model.Finding, failOn string) bool {
	if failOn == "" {
		return false
	}
	threshold := severityFromString(failOn)
	for _, finding := range findings {
		if model.SeverityRank(finding.Severity) >= model.SeverityRank(threshold) {
			return true
		}
	}
	return false
}

func severityFromString(value string) model.Severity {
	switch strings.ToLower(value) {
	case "critical":
		return model.SeverityCritical
	case "high":
		return model.SeverityHigh
	case "medium":
		return model.SeverityMedium
	case "low":
		return model.SeverityLow
	default:
		return model.SeverityCritical
	}
}

type parsedOptions struct {
	Values      map[string]string
	Positionals []string
}

func parseOptions(args []string, defaults map[string]string) (parsedOptions, error) {
	values := make(map[string]string, len(defaults))
	for key, value := range defaults {
		values[key] = value
	}
	var positionals []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			positionals = append(positionals, arg)
			continue
		}
		nameValue := strings.TrimPrefix(arg, "--")
		name, value, hasInlineValue := strings.Cut(nameValue, "=")
		if _, ok := values[name]; !ok {
			return parsedOptions{}, fmt.Errorf("unknown flag --%s", name)
		}
		if !hasInlineValue {
			if i+1 >= len(args) {
				return parsedOptions{}, fmt.Errorf("missing value for --%s", name)
			}
			i++
			value = args[i]
		}
		values[name] = value
	}
	return parsedOptions{Values: values, Positionals: positionals}, nil
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return home
}

func defaultTrustStorePath() string {
	return homeDir() + string(os.PathSeparator) + ".mcpcanary" + string(os.PathSeparator) + "trust.json"
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, `MCPCanary - local-first MCP trust scanner

Usage:
  mcpcanary scan [config] [--format text|json|markdown|sarif] [--fail-on high]
  mcpcanary inspect <config>
  mcpcanary lock <config> [--lockfile mcpcanary.lock]
  mcpcanary diff <config> [--lockfile mcpcanary.lock]
  mcpcanary version`)
}
