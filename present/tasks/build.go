package tasks

import (
	. "github.com/tbud/bud/context"
	"github.com/tbud/x/io/ioutil"
	"github.com/tbud/x/path/selector"
	"os"
	"path/filepath"

	"golang.org/x/tools/present"
)

type BuildPresentTask struct {
	commonCfg
	TargetDir string // build target dir

	target string // abs of target dir
}

func init() {
	buildTask := &BuildPresentTask{
		TargetDir: "present",
	}

	Task("build", PRESENT_TASK_GROUP, buildTask, Usage("Use to build static slides to target dir. Default target is 'present'."))
}

func (b *BuildPresentTask) Execute() (err error) {
	if err = initTemplates(b.RootTemplateDir); err != nil {
		return err
	}

	var s *selector.Selector
	if s, err = selector.New("*.(slide|article)"); err != nil {
		return err
	}

	var matches []string
	if matches, err = s.Matches("."); err != nil {
		return err
	}

	for _, m := range matches {
		filename := filepath.Base(m)
		slidePath := filepath.Join(b.target, filename+".html")
		var file *os.File
		if file, err = os.Create(slidePath); err != nil {
			return err
		}

		if err = renderDoc(file, m); err != nil {
			file.Close()
			return err
		}
		file.Close()
	}

	if err = ioutil.Copy(filepath.Join(b.TargetDir, "static"), filepath.Join(b.RootTemplateDir, "static"), 0, nil); err != nil {
		return err
	}

	return genPlayScriptFile(b.RootTemplateDir, "SocketTransport", filepath.Join(b.TargetDir, "static", "play.js"))
}

func (b *BuildPresentTask) Validate() (err error) {
	if err = b.commonCfg.Validate(); err != nil {
		return err
	}

	if filepath.IsAbs(b.TargetDir) {
		b.target = b.TargetDir
	} else {
		if b.target, err = filepath.Abs(b.TargetDir); err != nil {
			return err
		}
	}

	present.PlayEnabled = b.PlayEnabled

	return os.MkdirAll(b.target, 0744)
}
