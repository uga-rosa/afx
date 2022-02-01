package config

import (
	"bufio"
	"log"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// Config structure for file describing deployment. This includes the module source, inputs
// dependencies, backend etc. One config element is connected to a single deployment
type Config struct {
	GitHub []*GitHub `yaml:"github"`
	Gist   []*Gist   `yaml:"gist"`
	Local  []*Local  `yaml:"local"`
	HTTP   []*HTTP   `yaml:"http"`

	AppConfig *AppConfig `yaml:"config"`
}

// AppConfig represents configurations of this application itself
type AppConfig struct {
	Filter Filter `yaml:"filter"`
}

// Filter represents filter command. A filter command means command-line
// fuzzy finder, e.g. fzf
type Filter struct {
	Command string            `yaml:"command"`
	Args    []string          `yaml:"args"`
	Env     map[string]string `yaml:"env"`
}

// DefaultAppConfig is default settings of AppConfig
// Basically this will be overridden by user config if given
var DefaultAppConfig AppConfig = AppConfig{
	Filter: Filter{
		Command: "fzf",
		Args:    []string{"--ansi", "--no-preview", "--height=50%", "--reverse"},
	},
}

// Read reads yaml file based on given path
func Read(path string) (Config, error) {
	var cfg Config

	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	d := yaml.NewDecoder(
		bufio.NewReader(f),
		yaml.DisallowUnknownField(),
		yaml.DisallowDuplicateKey(),
	)
	if err := d.Decode(&cfg); err != nil {
		return cfg, err
	}

	return cfg, err
}

func parse(cfg Config) []Package {
	var pkgs []Package
	for _, pkg := range cfg.GitHub {
		// TODO: Remove?
		if pkg.HasReleaseBlock() && !pkg.HasCommandBlock() {
			pkg.Command = &Command{
				Link: []*Link{
					{From: filepath.Join("**", pkg.Release.Name)},
				},
			}
		}
		pkgs = append(pkgs, pkg)
	}
	for _, pkg := range cfg.Gist {
		pkgs = append(pkgs, pkg)
	}
	for _, pkg := range cfg.Local {
		pkgs = append(pkgs, pkg)
	}
	for _, pkg := range cfg.HTTP {
		pkgs = append(pkgs, pkg)
	}

	return pkgs
}

// Parse parses a config given via yaml files and converts it into package interface
func (c Config) Parse() ([]Package, error) {
	return parse(c), nil
}

func visitYAML(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		switch filepath.Ext(path) {
		case ".yaml", ".yml":
			*files = append(*files, path)
		}
		return nil
	}
}

// WalkDir walks given directory path and returns full-path of all yaml files
func WalkDir(path string) ([]string, error) {
	var files []string
	fi, err := os.Stat(path)
	if err != nil {
		return files, err
	}
	if fi.IsDir() {
		return files, filepath.Walk(path, visitYAML(&files))
	}
	switch filepath.Ext(path) {
	case ".yaml", ".yml":
		files = append(files, path)
	default:
		log.Printf("[WARN] %s: found but cannot be loaded. yaml is only allowed\n", path)
	}
	return files, nil
}
