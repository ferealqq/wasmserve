package pkg

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml"
)

var Config = new(config)

type config struct {
	TailwindExec   string   `toml:"tailwind_exec"`
	EnableTailwind bool     `toml:"enable_tailwind"`
	WasmFile       string   `toml:"wasm_file"`
	Root           string   `toml:"root"`
	TmpDir         string   `toml:"tmp_dir"`
	Build          cfgBuild `toml:"build"`

	WasmPath string
}

type cfgBuild struct {
	ExcludeDir  []string `toml:"exclude_dir"`
	ExcludeFile []string `toml:"exclude_file"`
}

func ReadConfig(path string) (*config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf := new(config)
	if err := toml.Unmarshal(data, conf); err != nil {
		return nil, err
	}
	conf.WasmPath = fmt.Sprintf("%s/%s", conf.TmpDir, conf.WasmFile)

	return conf, nil
}
