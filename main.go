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

package main

import (
	"bytes"
	"errors"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/cosmtrek/air/runner"
)

const indexHTML = `<!DOCTYPE html>
<!-- Polyfill for the old Edge browser -->
<script src="https://cdn.jsdelivr.net/npm/text-encoding@0.7.0/lib/encoding.min.js"></script>
<script src="wasm_exec.js"></script>
<script>
(async () => {
  const resp = await fetch('main.wasm');
  if (!resp.ok) {
    const pre = document.createElement('pre');
    pre.innerText = await resp.text();
    document.body.appendChild(pre);
  } else {
    const src = await resp.arrayBuffer();
    const go = new Go();
    const result = await WebAssembly.instantiate(src, go.importObject);
    go.argv = {{.Argv}};
    go.run(result.instance);
  }
})();
</script>
`

var (
	flagHTTP        = flag.String("http", ":8080", "HTTP bind address to serve")
	flagTags        = flag.String("tags", "", "Build tags")
	flagAllowOrigin = flag.String("allow-origin", "", "Allow specified origin (or * for all origins) to make requests to this server")
	flagOverlay     = flag.String("overlay", "", "Overwrite source files with a JSON file (see https://pkg.go.dev/cmd/go for more details)")
	flagTailwind    = flag.Bool("enable-tailwind", false, "Use tailwind")
	flagBuild       = flag.Bool("build", false, "Build tailwind & wasm")
	flagRun         = flag.Bool("run", false, "Run HTTP server")
	flagWatch       = flag.Bool("watch", false, "Watch file changes and serve http")
	flagAirConf     = flag.String("config", "air.toml", "Path to air.toml configuration")
)

var (
	waitChannel = make(chan struct{})
	conf        = new(config)
	cssFiles    = new(CssFiles)
)

func hasGo111Module(env []string) bool {
	for _, e := range env {
		if strings.HasPrefix(e, "GO111MODULE=") {
			return true
		}
	}
	return false
}

func handle(w http.ResponseWriter, r *http.Request) {
	if *flagAllowOrigin != "" {
		w.Header().Set("Access-Control-Allow-Origin", *flagAllowOrigin)
	}

	// output := conf.TmpDir

	upath := r.URL.Path[1:]
	fpath := path.Base(upath)
	// workdir := "."

	if !strings.HasSuffix(r.URL.Path, "/") {
		fi, err := os.Stat(fpath)
		if err != nil && !os.IsNotExist(err) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if fi != nil && fi.IsDir() {
			http.Redirect(w, r, r.URL.Path+"/", http.StatusSeeOther)
			return
		}
	}

	switch filepath.Base(fpath) {
	case ".":
		fpath = filepath.Join(fpath, "index.html")
		fallthrough
	case "index.html":
		if _, err := os.Stat(fpath); err != nil && !errors.Is(err, fs.ErrNotExist) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if errors.Is(err, fs.ErrNotExist) {
			fargs := flag.Args()
			argv := make([]string, 0, len(fargs))
			for _, a := range fargs {
				argv = append(argv, `"`+template.JSEscapeString(a)+`"`)
			}
			h := strings.ReplaceAll(indexHTML, "{{.Argv}}", "["+strings.Join(argv, ", ")+"]")
			http.ServeContent(w, r, "index.html", time.Now(), bytes.NewReader([]byte(h)))
			return
		}
	case "wasm_exec.js":
		if _, err := os.Stat(fpath); err != nil && !errors.Is(err, fs.ErrNotExist) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if errors.Is(err, fs.ErrNotExist) {
			out, err := exec.Command("go", "env", "GOROOT").Output()
			if err != nil {
				log.Print(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			f := filepath.Join(strings.TrimSpace(string(out)), "misc", "wasm", "wasm_exec.js")
			http.ServeFile(w, r, f)
			return
		}
	case conf.WasmFile:
		if _, err := os.Stat(fpath); err != nil && !errors.Is(err, fs.ErrNotExist) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if errors.Is(err, fs.ErrNotExist) {
			http.ServeFile(w, r, conf.WasmPath)
			return
		}
	}
	if *flagTailwind {
		if strings.HasSuffix(r.URL.Path, ".css") {
			out := cssFiles.getOutput(r.URL.Path)
			if out != "" {
				http.ServeFile(w, r, out)
			} else {
				http.Error(w, "css file not found", http.StatusInternalServerError)
			}
			return
		}
	}

	http.ServeFile(w, r, filepath.Join(".", r.URL.Path))
}

func main() {
	flag.Parse()

	if *flagBuild {
		c, err := readConfig(*flagAirConf)
		*conf = *c
		if err != nil {
			log.Fatal(http.ListenAndServe(*flagHTTP, nil))
			return
		}
		var wg sync.WaitGroup
		if *flagTailwind {
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
	} else if *flagRun {
		c, err := readConfig(*flagAirConf)
		*conf = *c
		if err != nil {
			log.Fatal(http.ListenAndServe(*flagHTTP, nil))
			return
		}

		// init css files from tmp
		initCssFiles()

		http.HandleFunc("/", handle)
		log.Fatal(http.ListenAndServe(*flagHTTP, nil))
	} else if *flagWatch {
		e, err := runner.NewEngine(*flagAirConf, true)
		if err != nil {
			log.Fatal(err.Error())
		}
		// Tell the user that build & bin should not be changed
		// Copy executable wasmserver to tmp
		e.Run()
		// Stop the process
	}
}
