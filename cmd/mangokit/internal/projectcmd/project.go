package projectcmd

import (
	"bytes"
	"context"
	"github.com/fatih/color"
	"github.com/mangohow/mangokit/cmd/mangokit/parallel"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ProjectGenerator struct {
	DirName     string
	ProjectName string
	BaseDir     string
	RepoUrl     string
	Branch      string
}

func (p *ProjectGenerator) Generate(ctx context.Context) error {
	// 克隆项目
	if err := p.CloneRepo(ctx); err != nil {
		return err
	}

	// 修改文件
	if err := p.ModifyFiles(); err != nil {
		// 清理文件夹
		os.RemoveAll(p.BaseDir)
		return err
	}

	// git init
	if err := p.GitInit(); err != nil {
		return err
	}

	color.Green("create project success!\n")
	color.Cyan("next step:\n")
	color.Cyan("cd %s && go mod tidy \n", p.DirName)

	return nil
}

func (p *ProjectGenerator) CloneRepo(ctx context.Context) error {
	// 克隆模板项目
	var cmd *exec.Cmd
	if p.Branch == "" {
		cmd = exec.CommandContext(ctx, "git", "clone", p.RepoUrl)
	} else {
		cmd = exec.CommandContext(context.Background(), "git", "clone", "-b", p.Branch, p.RepoUrl)
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		color.Red(string(output))
		color.Red("clone repo failed, %v\n", err)
		return err
	}

	words := strings.Split(p.RepoUrl, "/")
	oldDirName := strings.TrimSuffix(words[len(words)-1], ".git")

	// 修改目录名
	if err := os.Rename(oldDirName, p.DirName); err != nil {
		color.Red("modify dir name failed, %v\n", err)
		os.RemoveAll(oldDirName)
		return err
	}

	return nil
}

func (p *ProjectGenerator) ModifyFiles() error {
	// 1.删除.git文件夹
	gitDir := filepath.Join(p.BaseDir, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		color.Red("remove .git failed, %v\n", err)
		return err
	}

	// 2. 修改Go mod中的module, 并返回之前的名称
	module, err := p.ModifyGoModule()
	if err != nil {
		color.Red("modify go.mod failed, %v\n", err)
		return err
	}

	// 3. 修改go文件中的import
	if err := p.ModifyGoImports(module); err != nil {
		return err
	}

	return nil
}

func (p *ProjectGenerator) ModifyGoModule() (string, error) {
	goModPath := filepath.Join(p.BaseDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		color.Red("read go.mod failed, %v\n", err)
		return "", err
	}

	// 获取原module名称
	index := bytes.Index(content, []byte("module")) + len("module")
	lineEndIndex := bytes.Index(content, []byte("\n"))
	moduleStr := string(content[index:lineEndIndex])
	module := strings.Trim(moduleStr, " \r\n")

	// 写入新的名称
	f, err := os.OpenFile(goModPath, os.O_RDWR, 0644)
	if err != nil {
		color.Red("open go.mod failed, %v\n", err)
		return "", err
	}
	defer f.Close()
	if err := f.Truncate(0); err != nil {
		color.Red("truncate go.mod failed, %v\n", err)
		return "", err
	}
	if _, err := f.Write([]byte("module " + p.ProjectName + "\n")); err != nil {
		color.Red("write go.mod failed, %v\n", err)
		return "", err
	}
	if _, err := f.Write(content[lineEndIndex+1:]); err != nil {
		color.Red("write go.mod failed, %v\n", err)
		return "", err
	}

	return module, nil
}

func (p *ProjectGenerator) ModifyGoImports(oldModule string) error {
	goFiles := make([]string, 0)

	// 获取项目目录下所有go文件
	err := filepath.WalkDir(p.BaseDir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".go") {
			goFiles = append(goFiles, path)
		}

		return nil
	})
	if err != nil {
		color.Red("walk dir failed, %v\n", err)
		return err
	}

	// 修改go imports
	oldMod := []byte(oldModule)
	newMod := []byte(p.ProjectName)

	// 启动多个goroutine来修改go文件中的import
	err = parallel.Parallel(context.Background(), 8, goFiles, func(file string) error {
		return p.modifyGoImports(file, oldMod, newMod)
	})
	if err != nil {
		color.Red("modify go imports error, %v\n", err)
		return err
	}

	return nil
}

func (p *ProjectGenerator) modifyGoImports(path string, oldModule, newModule []byte) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	replaced := bytes.ReplaceAll(content, oldModule, newModule)

	if err = os.WriteFile(path, replaced, 0644); err != nil {
		return err
	}

	return nil
}

func (p *ProjectGenerator) GoModInit() error {
	cmd := exec.Command("go", "mod", "init", p.ProjectName)
	// 进入base dir
	if err := os.Chdir(p.BaseDir); err != nil {
		color.Red("change dir failed, %v\n", err)
		return err
	}
	if err := cmd.Run(); err != nil {
		color.Red("go mod init failed, %v\n", err)
		return err
	}
	os.Chdir("..")

	return nil
}

func (p *ProjectGenerator) GitInit() error {
	cmd := exec.Command("git", "init")
	// 进入base dir
	if err := os.Chdir(p.BaseDir); err != nil {
		color.Red("change dir failed, %v\n", err)
		return err
	}
	if err := cmd.Run(); err != nil {
		color.Red("git init failed, %v\n", err)
		return err
	}
	os.Chdir("..")

	return nil
}
