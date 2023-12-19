package projectcmd

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fatih/color"
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
	oldDirName := words[len(words)-1]

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

	// 2.从 go mod中获取原项目名
	module, err := p.GetGoModName()
	if err != nil {
		return err
	}

	// 3. 删除go.mod
	if err := os.Remove(filepath.Join(p.BaseDir, "go.mod")); err != nil {
		color.Red("remove go.mod failed, %v\n", err)
		return err
	}

	// 4. 修改go文件中的import
	if err := p.ModifyGoImports(module); err != nil {
		return err
	}

	// 5. 重新初始化go mod
	return p.GoModInit()
}

func (p *ProjectGenerator) GetGoModName() (string, error) {
	goModPath := filepath.Join(p.BaseDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		color.Red("read go.mod failed, %v\n", err)
		return "", err
	}

	index := bytes.Index(content, []byte("module")) + len("module")
	lineEndIndex := bytes.Index(content, []byte("\n"))
	moduleStr := string(content[index:lineEndIndex])
	module := strings.Trim(moduleStr, " \r\n")

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

	ch := make(chan string, 1024)
	errCh := make(chan error, 1)
	group := sync.WaitGroup{}
	counter := make(chan struct{}, 1024)
	ctx, cancel := context.WithCancel(context.Background())

	// 修改go imports
	oldMod := []byte(oldModule)
	newMod := []byte(p.ProjectName)
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()

			for {
				select {
				case file := <-ch:
					if err := p.modifyGoImports(file, oldMod, newMod); err != nil {
						errCh <- err
						return
					}
					counter <- struct{}{}
				case <-ctx.Done():
					return
				}
			}

		}()
	}

	for i := 0; i < len(goFiles); i++ {
		ch <- goFiles[i]
	}

	// 等待任务完成
	count := 0
loop:
	for {
		select {
		case err = <-errCh:
			color.Red("modify go imports error, %v\n", err)
			break loop
		case <-counter:
			count++
			if count == len(goFiles) {
				break loop
			}
		}
	}

	cancel()
	group.Wait()

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
