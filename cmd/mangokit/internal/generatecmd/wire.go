package generatecmd

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/mangohow/mangokit/cmd/mangokit/parallel"
	"github.com/spf13/cobra"
)

var CmdGenWire = &cobra.Command{
	Use:   "wire",
	Short: "generate wire injection files",
	Long:  "generate wire injection files",
	Run: func(cmd *cobra.Command, args []string) {
		GenerateWire()
	},
}

func GenerateWire()  error{
	c := exec.Command("go", "mod", "tidy")
	if err := c.Run(); err != nil {
		return err
		color.Red("go mod tidy, %v\n", err)
	}

	path, err := os.Getwd()
	if err != nil {
		color.Red("get current working directory failed, %v", err)
		return err
	}

	goFiles := make([]string, 0)
	// 获取所有go文件
	err = filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".go") {
			goFiles = append(goFiles, path)
		}

		return nil
	})
	if err != nil {
		color.Red("walk go files error, %v", err)
		return err
	}

	args := []string{"generate"}

	// 启动多个goroutine统计文件中是否有//go:generate
	resChan := make(chan string, 1024)
	res, err := parallel.ParallelResult(context.Background(), 8, goFiles, resChan, func(path string) error {
		// 最多读取2K
		content, err := readFile(path, 2048)
		if err != nil {
			color.Red("read file %v error, %v", path, err)
			return err
		}

		// 判断文件中是否有//go:generate
		if bytes.Contains(content, []byte("//go:generate")) {
			resChan <- path
		}

		return nil
	})
	if err != nil {
		return err
	}
	args = append(args, res...)

	command := exec.Command("go", args...)
	output, err := command.CombinedOutput()
	if err != nil {
		color.Red("go generate error, %v\n", err)
		return err
	}
	fmt.Println(string(output))

	return nil
}

func readFile(name string, readN int) ([]byte, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	content := make([]byte, readN)
	n, err := f.Read(content)
	if err != nil {
		return nil, err
	}

	return content[:n], nil
}
