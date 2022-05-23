// Copyright 2018 Hajime Hoshi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	. "github.com/hajimehoshi/wasmserve/pkg"
	"github.com/spf13/cobra"
)

// TODO Move to root?
var cssFiles = new(CssFiles)

type excludableDirs []string

func (e *excludableDirs) Contains(dir string) bool {
	for _, n := range *e {
		if n == dir {
			return true
		}
	}

	return false
}

func removeIfContains(e []string, dir string) []string {
	for i, n := range e {
		if n == dir {
			return append(e[:i], e[i+1:]...)
		}
	}

	return e
}

func buildTailwindCss(cssPath string) (*CssPath, error) {
	output := Config.TmpDir

	rf := strings.Split(cssPath, "/")
	filename := rf[len(rf)-1]

	workdir := "."
	outpath := filepath.Join(output, filename)
	args := []string{"-i", cssPath, "-o", outpath}

	var ex string
	// if the user is using npx tailwindcss or something like that we need to seperate the starting point and the rest
	if rest := strings.Split(Config.TailwindExec, " "); len(rest) > 1 {
		ex = rest[0]
		log.Println(rest)
		rest = rest[1:]
		args = append(rest, args...)
	} else {
		ex = Config.TailwindExec
	}

	cmdBuild := exec.Command(ex, args...)
	cmdBuild.Dir = workdir
	out, err := cmdBuild.CombinedOutput()
	if err != nil {
		log.Print(err)
		log.Print(string(out))
		return nil, err
	}

	return &CssPath{Output: outpath, Input: cssPath}, nil
}

func cssFilesFromDir(rd string) []string {
	var excludeDirs excludableDirs = Config.Build.ExcludeDir
	// Remove if
	excludeDirs = removeIfContains(excludeDirs, rd)
	var compilable []string
	err := filepath.Walk(rd, func(path string, info os.FileInfo, err error) error {
		// path is absolute
		if err != nil {
			return err
		}
		// jos se tmp on excludeDirs niin älä excludee
		if excludeDirs.Contains(info.Name()) {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".css") {
			compilable = append(compilable, path)
		}

		return nil
	})
	if err != nil {
		log.Print(err.Error())
		return nil
	}

	return compilable
}

func buildAllCssFiles() {
	var compilable = cssFilesFromDir(".")
	var wg sync.WaitGroup

	for _, f := range compilable {
		wg.Add(1)

		go func(file string) {
			defer wg.Done()
			cssPath, err := buildTailwindCss(file)
			if err == nil {
				cssFiles.Add(cssPath)
			} else {
				log.Print(err.Error())
			}
		}(f)
	}

	wg.Wait()
}

func initCssFiles() {
	files := cssFilesFromDir(Config.TmpDir)

	for _, f := range files {
		cssFiles.Add(&CssPath{Output: f})
	}
}

func hasGo111Module(env []string) bool {
	for _, e := range env {
		if strings.HasPrefix(e, "GO111MODULE=") {
			return true
		}
	}
	return false
}

func buildWasm() {
	// go build
	args := []string{"build", "-o", Config.WasmPath}
	// move flags to conf
	// if *flagTags != "" {
	// 	args = append(args, "-tags", *flagTags)
	// }
	// if *flagOverlay != "" {
	// 	args = append(args, "-overlay", *flagOverlay)
	// }
	// if len(flag.Args()) > 0 {
	// 	args = append(args, flag.Args()[0])
	// } else {
	// 	args = append(args, ".")
	// }
	args = append(args, ".")

	cmdBuild := exec.Command("go", args...)
	cmdBuild.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	// If GO111MODULE is not specified explicitly, enable Go modules.
	// Enabling this is for backward compatibility of wasmserve.
	if !hasGo111Module(cmdBuild.Env) {
		cmdBuild.Env = append(cmdBuild.Env, "GO111MODULE=on")
	}
	cmdBuild.Dir = Config.Root
	out, err := cmdBuild.CombinedOutput()
	if err != nil {
		log.Print(err)
		log.Print(string(out))
		return
	}
	if len(out) > 0 {
		log.Print(string(out))
	}
}

var buildCmd = &cobra.Command{
	Use:              "build",
	Short:            "build all the webassembly and css files",
	Long:             `TODO`,
	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println(flagConf)
		if err := initConf(); err != nil {
			log.Fatal(err)
			return
		}

		var wg sync.WaitGroup
		if Config.EnableTailwind {
			wg.Add(1)
			go func() {
				defer wg.Done()
				buildAllCssFiles()
			}()
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			buildWasm()
		}()

		wg.Wait()
	},
}
