package cmd

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
	"text/template"
	"time"

	. "github.com/hajimehoshi/wasmserve/pkg"
	"github.com/spf13/cobra"
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

func handle(w http.ResponseWriter, r *http.Request) {
	// TODO Move to Config
	// if *flagAllowOrigin != "" {
	// 	w.Header().Set("Access-Control-Allow-Origin", *flagAllowOrigin)
	// }

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
	case Config.WasmFile:
		if _, err := os.Stat(fpath); err != nil && !errors.Is(err, fs.ErrNotExist) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if errors.Is(err, fs.ErrNotExist) {
			http.ServeFile(w, r, Config.WasmPath)
			return
		}
	}
	if Config.EnableTailwind {
		if strings.HasSuffix(r.URL.Path, ".css") {
			out := cssFiles.GetOutput(r.URL.Path)
			if out != "" {
				http.ServeFile(w, r, out)
			} else {
				http.Error(w, "css file not found", http.StatusInternalServerError)
			}
			return
		}
	}

	if f, err := os.Stat(filepath.Join(".", r.URL.Path)); errors.Is(err, os.ErrNotExist){
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
	}else{
		http.ServeFile(w, r, f.Name())
	}
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run HTTP server that serves the built webassembly and other static files",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := initConf(); err != nil {
			log.Fatal(err)
			return
		}

		initCssFiles()

		http.HandleFunc("/", handle)
		port := ":"+Config.Http
		log.Printf("Trying to listen to port: "+port)
		log.Printf("Listening connections on http://localhost%s", port)
		log.Fatal(http.ListenAndServe(port, nil))
	},
}
