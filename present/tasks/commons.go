package tasks

import (
	"go/build"
	"path/filepath"

	. "github.com/tbud/bud/context"
	. "github.com/tbud/x/config"
)

const (
	PRESENT_TASK_GROUP_NAME = "present"
	PRESENT_TASK_GROUP      = Group(PRESENT_TASK_GROUP_NAME)
	PRESENT_BASE_PKG        = "github.com/tbud/seeds/present"
)

type commonCfg struct {
	BaseDir         string // base path for slide theme template and static resources
	Theme           string // theme name with base dir or theme absolutely path
	RootTemplateDir string // template root path
	PlayEnabled     bool   // enable playground (permit execution of arbitrary user code)
}

func (c *commonCfg) Validate() error {
	if len(c.BaseDir) == 0 {
		p, err := build.Import(PRESENT_BASE_PKG, "", build.FindOnly)
		if err != nil {
			Log.Error("Couldn't find go present files: %v\n", err)
			return err
		}
		c.BaseDir = p.Dir
	}

	if filepath.IsAbs(c.Theme) {
		c.RootTemplateDir = c.Theme
	} else {
		c.RootTemplateDir = filepath.Join(c.BaseDir, "themes", c.Theme)
	}

	return nil
}

func init() {
	TaskConfig(PRESENT_TASK_GROUP_NAME, Config{
		"baseDir":     "",
		"theme":       "default",
		"playEnabled": true,
	})
}
