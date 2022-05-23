package pkg

import (
	"fmt"
	"os"
	"time"

	"github.com/pelletier/go-toml"
)

var Config = new(config)

var (
	DefaultTomlFile = "wasmserve.toml"
	DefaultHttp     = "8080"
	DefaultTags     = ""
	// TODO Check how the toml unmarshal handles this
	DefaultAllowOrigin = ""
	DefaultOverlay     = ""
	DefaultWasmFile    = "main.wasm"
	DefaultTmpDir      = "tmp"
	DefaultRoot        = "."
)

type config struct {
	UseAir         bool   `toml:"use_air"`
	TailwindExec   string `toml:"tailwind_exec,omitempty"`
	EnableTailwind bool   `toml:"enable_tailwind"`
	WasmFile       string `toml:"wasm_file,omitempty"`
	Http           string `toml:"http,omitempty"`
	Tags           string `toml:"tags,omitempty"`
	AllowOrigin    string `toml:"allow_origin,omitempty"`
	Overlay        string `toml:"overlay,omitempty"`
	// Air configs
	Root        string    `toml:"root"`
	TmpDir      string    `toml:"tmp_dir"`
	TestDataDir string    `toml:"testdata_dir,omitempty"`
	Build       cfgBuild  `toml:"build"`
	Color       cfgColor  `toml:"color"`
	Log         cfgLog    `toml:"log"`
	Misc        cfgMisc   `toml:"misc"`
	Screen      cfgScreen `toml:"screen"`

	WasmPath string `commented:"true"`
}

type cfgBuild struct {
	Cmd              string        `toml:"cmd"`
	Bin              string        `toml:"bin"`
	FullBin          string        `toml:"full_bin"`
	ArgsBin          []string      `toml:"args_bin"`
	Log              string        `toml:"log"`
	IncludeExt       []string      `toml:"include_ext"`
	ExcludeDir       []string      `toml:"exclude_dir"`
	IncludeDir       []string      `toml:"include_dir"`
	ExcludeFile      []string      `toml:"exclude_file"`
	ExcludeRegex     []string      `toml:"exclude_regex"`
	ExcludeUnchanged bool          `toml:"exclude_unchanged"`
	FollowSymlink    bool          `toml:"follow_symlink"`
	Delay            int           `toml:"delay"`
	StopOnError      bool          `toml:"stop_on_error"`
	SendInterrupt    bool          `toml:"send_interrupt"`
	KillDelay        time.Duration `toml:"kill_delay"`
}

type cfgColor struct {
	Main    string `toml:"main"`
	Watcher string `toml:"watcher"`
	Build   string `toml:"build"`
	Runner  string `toml:"runner"`
	App     string `toml:"app"`
}

type cfgLog struct {
	AddTime bool `toml:"time"`
}

type cfgMisc struct {
	CleanOnExit bool `toml:"clean_on_exit"`
}

type cfgScreen struct {
	ClearOnRebuild bool `toml:"clear_on_rebuild"`
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

func DefaultConfig() config {
	return config{
		UseAir:         false,
		EnableTailwind: false,
		WasmFile:       DefaultWasmFile,
		Http:           DefaultHttp,
		Tags:           DefaultTags,
		AllowOrigin:    DefaultAllowOrigin,
		Overlay:        DefaultOverlay,
		Root:           DefaultRoot,
		TmpDir:         DefaultTmpDir,
		WasmPath:       fmt.Sprintf("%s/%s", DefaultTmpDir, DefaultWasmFile),
	}
}

func DefaultTomlContent() config {
	return config{
		UseAir:         true,
		EnableTailwind: false,
		WasmFile:       DefaultWasmFile,
		Http:           DefaultHttp,
		Tags:           DefaultTags,
		AllowOrigin:    DefaultAllowOrigin,
		Overlay:        DefaultOverlay,
		Root:           DefaultRoot,
		TmpDir:         DefaultTmpDir,
		Build: cfgBuild{
			Cmd:              "wasmserve build",
			Bin:              "wasmserve run",
			FullBin:          "wasmserve run",
			IncludeExt:       []string{"go", "tpl", "tmpl", "html", "css"},
			ExcludeDir:       []string{"assets", "tmp", "vendor", "frontend/node_modules"},
			IncludeDir:       []string{},
			ExcludeFile:      []string{},
			ExcludeRegex:     []string{"_test.go"},
			ExcludeUnchanged: true,
			FollowSymlink:    true,
			Log:              "air.log",
			Delay:            1000,
			StopOnError:      true,
			SendInterrupt:    true,
			KillDelay:        400,
			ArgsBin:          []string{},
		},
		Log: cfgLog{
			AddTime: false,
		},
		Color: cfgColor{
			Main:    "magenta",
			Watcher: "cyan",
			Build:   "yellow",
			Runner:  "green",
		},
		Misc: cfgMisc{
			CleanOnExit: false,
		},
	}
}
