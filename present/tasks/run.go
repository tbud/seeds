package tasks

import (
	. "github.com/tbud/bud/context"
	"go/build"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/tools/playground/socket"
	"golang.org/x/tools/present"
)

type RunPresentTask struct {
	HttpAddr     string // HTTP service address (e.g., '127.0.0.1:3999')
	OriginHost   string // host component of web origin URL (e.g., 'localhost')
	BaseDir      string // base path for slide theme template and static resources
	Theme        string // theme name with base dir or theme absolutely path
	PlayEnabled  bool   // enable playground (permit execution of arbitrary user code)
	NativeClient bool   // use Native Client environment playground (prevents non-Go code execution)

	rootDir string // template root path
}

func init() {
	runTask := &RunPresentTask{
		HttpAddr:     "127.0.0.1:3999",
		OriginHost:   "",
		Theme:        "default",
		PlayEnabled:  true,
		NativeClient: false,
	}

	Task("run", PRESENT_TASK_GROUP, runTask, Usage("Command to run present server."))
}

func (r *RunPresentTask) Execute() error {
	ln, err := net.Listen("tcp", r.HttpAddr)
	if err != nil {
		Log.Fatal("%v", err)
		return err
	}
	defer ln.Close()

	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		Log.Fatal("%v", err)
		return err
	}

	origin := &url.URL{Scheme: "http"}
	if len(r.OriginHost) > 0 {
		origin.Host = net.JoinHostPort(r.OriginHost, port)
	} else if ln.Addr().(*net.TCPAddr).IP.IsUnspecified() {
		host, _ := os.Hostname()
		origin.Host = net.JoinHostPort(host, port)
	} else {
		reqHost, reqPort, err := net.SplitHostPort(r.HttpAddr)
		if err != nil {
			Log.Fatal("%v", err)
			return err
		}
		if reqPort == "0" {
			origin.Host = net.JoinHostPort(reqHost, port)
		} else {
			origin.Host = r.HttpAddr
		}
	}

	if r.PlayEnabled {
		if r.NativeClient {
			socket.RunScripts = false
			socket.Environ = func() []string {
				if runtime.GOARCH == "amd64" {
					return environ("GOOS=nacl", "GOARCH=amd64p32")
				}
				return environ("GOOS=nacl")
			}
		}
		playScript(r.rootDir, "SocketTransport")
		http.Handle("/socket", socket.NewHandler(origin))
	}
	http.Handle("/static/", http.FileServer(http.Dir(r.rootDir)))

	if !ln.Addr().(*net.TCPAddr).IP.IsLoopback() &&
		r.PlayEnabled && !r.NativeClient {
		Log.Error(localhostWarning)
	}

	Log.Info("Open your web browser and visit %s", origin.String())
	return http.Serve(ln, nil)
}

func (r *RunPresentTask) Validate() error {
	if len(r.BaseDir) == 0 {
		p, err := build.Import(PRESENT_BASE_PKG, "", build.FindOnly)
		if err != nil {
			Log.Error("Couldn't find go present files: %v\n", err)
			return err
		}
		r.BaseDir = p.Dir
	}

	present.PlayEnabled = r.PlayEnabled

	if filepath.IsAbs(r.Theme) {
		r.rootDir = r.Theme
	} else {
		r.rootDir = filepath.Join(r.BaseDir, "themes", r.Theme)
	}

	err := initTemplates(r.rootDir)
	if err != nil {
		Log.Fatal("Failed to parse templates: %v", err)
		return err
	}

	return nil
}

func environ(vars ...string) []string {
	env := os.Environ()
	for _, r := range vars {
		k := strings.SplitAfter(r, "=")[0]
		var found bool
		for i, v := range env {
			if strings.HasPrefix(v, k) {
				env[i] = r
				found = true
			}
		}
		if !found {
			env = append(env, r)
		}
	}
	return env
}

const basePathMessage = `
By default, gopresent locates the slide template files and associated
static content by looking for a %q package
in your Go workspaces (GOPATH).

You may use the -base flag to specify an alternate location.
`

const localhostWarning = `
WARNING!  WARNING!  WARNING!

The present server appears to be listening on an address that is not localhost.
Anyone with access to this address and port will have access to this machine as
the user running present.

To avoid this message, listen on localhost or run with -play=false.

If you don't understand this message, hit Control-C to terminate this process.

WARNING!  WARNING!  WARNING!
`
