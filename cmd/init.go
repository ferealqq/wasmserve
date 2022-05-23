package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"

	. "github.com/hajimehoshi/wasmserve/pkg"
)

const defaultTailwindConfig = `module.exports = {
	content: ["./**/*.{html,go}"],
	theme: {
	  extend: {},
	},
	plugins: [],
}`

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
	enableTailwind := true
	prompt := &survey.Confirm{Message: "Would you like to enable tailwind?"}
	survey.AskOne(prompt, &enableTailwind)

	tomlConfig := DefaultTomlContent()
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
		tomlConfig.EnableTailwind = true
		switch tailwindExec {
		case NPX:
			tomlConfig.TailwindExec = "npx tailwindcss"
		case MANUAL:
			tomlConfig.TailwindExec = "CUSTOM TAILWIND BUILD"
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

			err = os.Chmod(tailfile, 0755)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("Tailwind executable downloaded")

			tomlConfig.TailwindExec = fmt.Sprintf("./%s", tailfile)
		}
		defaultTailwind := true
		useDefault := &survey.Confirm{Message: "Would you like to use the default tailwind.config.js?"}
		survey.AskOne(useDefault, &enableTailwind)
		if defaultTailwind {
			e := os.WriteFile("tailwind.config.js", []byte(defaultTailwindConfig), 0644)
			if e != nil {
				log.Fatal(e)
			}
			fmt.Println("tailwind.config.js created")
		}
	}
	b, err := toml.Marshal(tomlConfig)
	if err != nil {
		log.Fatal(err)
		return
	}

	if err := os.WriteFile("wasmserve.toml", b, 0644); err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("\nwasmserve.toml created")

	fmt.Printf("Start your wasm server with: \n\twasmserve --watch\n\n")

	fmt.Println("Happy hacking :)")
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "initialize wasm configuration file",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		__init()
	},
}
