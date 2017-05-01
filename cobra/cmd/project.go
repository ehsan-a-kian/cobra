package cmd

import (
	"os"
	"path/filepath"
	"strings"
)

// Project contains name, license and paths to projects.
type Project struct {
	absPath string
	cmdPath string
	srcPath string
	license License
	name    string
}

// NewProject returns Project with specified project name.
// If projectName is blank string, it returns nil.
func NewProject(projectName string) *Project {
	if projectName == "" {
		return nil
	}

	p := new(Project)
	p.name = projectName

	// 1. Find already created protect.
	p.absPath = findPackage(projectName)

	// 2. If there are no created project with this path, and user is in GOPATH,
	// then use GOPATH/src+projectName.
	if p.absPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			er(err)
		}
		for _, goPath := range goPaths {
			if filepath.HasPrefix(wd, goPath) {
				p.absPath = filepath.Join(goPath, "src", projectName)
				break
			}
		}
	}

	// 3. If user is not in GOPATH, then use (first GOPATH)+projectName.
	if p.absPath == "" {
		p.absPath = filepath.Join(srcPaths[0], projectName)
	}

	return p
}

// findPackage returns full path to existing go package in GOPATHs.
// findPackage returns "", if it can't find path.
// If packageName is "", findPackage returns "".
func findPackage(packageName string) string {
	if packageName == "" {
		return ""
	}

	for _, srcPath := range srcPaths {
		packagePath := filepath.Join(srcPath, packageName)
		if exists(packagePath) {
			return packagePath
		}
	}

	return ""
}

// NewProjectFromPath returns Project with specified absolute path to
// package.
// If absPath is blank string or if absPath is not actually absolute,
// it returns nil.
func NewProjectFromPath(absPath string) *Project {
	if absPath == "" || !filepath.IsAbs(absPath) {
		return nil
	}

	p := new(Project)
	p.absPath = absPath
	p.absPath = strings.TrimSuffix(p.absPath, findCmdDir(p.absPath))
	p.name = filepath.ToSlash(trimSrcPath(p.absPath, p.SrcPath()))
	return p
}

// trimSrcPath trims at the beginning of absPath the srcPath.
func trimSrcPath(absPath, srcPath string) string {
	relPath, err := filepath.Rel(srcPath, absPath)
	if err != nil {
		er("Cobra supports project only within $GOPATH")
	}
	return relPath
}

// License returns the License object of project.
func (p *Project) License() License {
	if p.license.Text == "" && p.license.Name != "None" {
		p.license = getLicense()
	}

	return p.license
}

// Name returns the name of project, e.g. "github.com/spf13/cobra"
func (p Project) Name() string {
	return p.name
}

// CmdPath returns absolute path to directory, where all commands are located.
//
// CmdPath returns blank string, only if p.AbsPath() is a blank string.
func (p *Project) CmdPath() string {
	if p.absPath == "" {
		return ""
	}
	if p.cmdPath == "" {
		p.cmdPath = filepath.Join(p.absPath, findCmdDir(p.absPath))
	}
	return p.cmdPath
}

// findCmdDir checks if base of absPath is cmd dir and returns it or
// looks for existing cmd dir in absPath.
// If the cmd dir doesn't exist, empty, or cannot be found,
// it returns "cmd".
func findCmdDir(absPath string) string {
	if !exists(absPath) || isEmpty(absPath) {
		return "cmd"
	}

	base := filepath.Base(absPath)
	for _, cmdDir := range cmdDirs {
		if base == cmdDir {
			return cmdDir
		}
	}

	files, _ := filepath.Glob(filepath.Join(absPath, "c*"))
	for _, file := range files {
		for _, cmdDir := range cmdDirs {
			if file == cmdDir {
				return cmdDir
			}
		}
	}

	return "cmd"
}

// AbsPath returns absolute path of project.
func (p Project) AbsPath() string {
	return p.absPath
}

// SrcPath returns absolute path to $GOPATH/src where project is located.
func (p *Project) SrcPath() string {
	if p.srcPath != "" {
		return p.srcPath
	}
	if p.absPath == "" {
		p.srcPath = srcPaths[0]
		return p.srcPath
	}

	for _, srcPath := range srcPaths {
		if strings.HasPrefix(p.absPath, srcPath) {
			p.srcPath = srcPath
			break
		}
	}

	return p.srcPath
}