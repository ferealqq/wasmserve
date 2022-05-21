package main

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
)

type config struct {
	Root     string   `toml:"root"`
	TmpDir   string   `toml:"tmp_dir"`
	Build    cfgBuild `toml:"build"`
	wasmPath string
}

type cfgBuild struct {
	ExcludeDir  []string `toml:"exclude_dir"`
	ExcludeFile []string `toml:"exclude_file"`
}

func readConfig(path string) (*config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf := new(config)
	if err := toml.Unmarshal(data, conf); err != nil {
		return nil, err
	}
	conf.wasmPath = filepath.Join(conf.TmpDir, "main.wasm")

	return conf, nil
}
