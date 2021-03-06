package tasks

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/tools/godoc/static"
)

var scripts = []string{"jquery.js", "jquery-ui.js", "playground.js", "play.js"}

// playScript registers an HTTP handler at /play.js that serves all the
// scripts specified by the variable above, and appends a line that
// initializes the playground with the specified transport.
func playScript(root, transport string) {
	modTime := time.Now()
	buf := mergeJs(root, transport)
	http.HandleFunc("/static/play.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/javascript")
		http.ServeContent(w, r, "", modTime, bytes.NewReader(buf))
	})
}

func genPlayScriptFile(root, transport, scriptFile string) error {
	if file, err := os.Create(scriptFile); err != nil {
		return err
	} else {
		file.Write(mergeJs(root, transport))
		file.Close()
	}
	return nil
}

func mergeJs(root, transport string) []byte {
	var buf bytes.Buffer
	for _, p := range scripts {
		if s, ok := static.Files[p]; ok {
			buf.WriteString(s)
			continue
		}
		b, err := ioutil.ReadFile(filepath.Join(root, "static", p))
		if err != nil {
			panic(err)
		}
		buf.Write(b)
	}
	fmt.Fprintf(&buf, "\ninitPlayground(new %v());\n", transport)
	return buf.Bytes()
}
