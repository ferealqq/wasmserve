package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

func createConfTomlString(enableTailwind bool, tailwindExec string) string {
	return fmt.Sprintf(`
# Wasm serve config
enable_tailwind=%s
tailwind_exec="%s"
wasm_file="main.wasm"

# AIR Config
# Config file for [Air](https://github.com/cosmtrek/air) in TOML format
# Working directory
# . or absolute path, please note that the directories following must be under root.
root = "."
tmp_dir = "tmp"

[build]
# Just plain old shell command. You could use 'make' as well.
cmd = "wasmserve --build --enable-tailwind"
# Binary file yields from 'cmd'.
bin = "wasmserve --run --enable-tailwind"
# Customize binary, can setup environment variables when run your app.
full_bin = "wasmserve --run --enable-tailwind"
# Watch these filename extensions.
include_ext = ["go", "tpl", "tmpl", "html", "css"]
# Ignore these filename extensions or directories.
exclude_dir = ["assets", "tmp", "vendor", "frontend/node_modules"]
# Watch these directories if you specified.
include_dir = []
# Exclude files.
exclude_file = []
# Exclude specific regular expressions.
exclude_regex = ["_test.go"]
# Exclude unchanged files.
exclude_unchanged = true
# Follow symlink for directories
follow_symlink = true
# This log file places in your tmp_dir.
log = "air.log"
# It's not necessary to trigger build each time file changes if it's too frequent.
delay = 1000 # ms
# Stop running old binary when build errors occur.
stop_on_error = true
# Send Interrupt signal before killing process (windows does not support this feature)
send_interrupt = true
# Delay after sending Interrupt signal
kill_delay = 400 # ms
# Add additional arguments when running binary (bin/full_bin). Will run './tmp/main hello world'.
args_bin = []

[log]
# Show log time
time = false

[color]
# Customize each part's color. If no color found, use the raw app log.
main = "magenta"
watcher = "cyan"
build = "yellow"
runner = "green"

[misc]
# Delete tmp directory on exit
clean_on_exit = true
`, strconv.FormatBool(enableTailwind), tailwindExec)
}

func getOsName() string {
	switch runtime.GOOS {
	case "darwin":
		return "macos"
	case "linux":
		return "linux"
	default:
		return "windows"
	}
}

func getArchitecture() string {
	// Tailwindcss executable only supports x64 and arm64
	if strings.Contains(runtime.GOARCH, "arm") {
		return "arm64"
	} else {
		return "x64"
	}
}

func __init() {
	enableTailwind := false
	var confToml string
	prompt := &survey.Confirm{Message: "Would you like to enable tailwind?"}
	survey.AskOne(prompt, &enableTailwind)

	if enableTailwind {
		var (
			NPX    = "Use npx tailwindcss"
			LOCAL  = "Download tailwindcss executable to local directory"
			MANUAL = "Manual"
		)
		var tailwindExec string
		prompt := &survey.Select{
			Message: "Which tailwindcss build option would you like to use:",
			Options: []string{NPX, LOCAL, MANUAL},
		}
		survey.AskOne(prompt, &tailwindExec)
		switch tailwindExec {
		case NPX:
			confToml = createConfTomlString(true, "npx tailwindcss")
		case MANUAL:
			confToml = createConfTomlString(true, "CUSTOM TAILWIND BUILD COMMAND")
			fmt.Printf("Remember to specify your tailwind build command in wasmserve.toml\n")
		case LOCAL:
			url := fmt.Sprintf(
				"https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-%s-%s",
				getOsName(),
				getArchitecture(),
			)
			fmt.Printf("Downloading: %s\n", url)

			tailfile := "tailwindcss"
			resp, err := http.Get(url)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			out, err := os.Create(tailfile)
			if err != nil {
				log.Fatal(err)
			}

			defer out.Close()

			_, err = io.Copy(out, resp.Body)

			if err != nil {
				log.Fatal(err)
			}

			err = os.Chmod(tailfile, 0777)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("Tailwind executable downloaded")

			confToml = createConfTomlString(true, "./tailwindcss")
		}
	} else {
		confToml = createConfTomlString(false, "")
	}
	err := os.WriteFile("wasmserve.toml", []byte(confToml), 0755)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nwasmserve.toml created")

	fmt.Printf("Start your wasm server with: \n\twasmserve --watch\n\n")

	fmt.Println("Happy hacking :)")
}
