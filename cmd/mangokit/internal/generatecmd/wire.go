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
	"sync"

	"github.com/fatih/color"
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
	taskCh := make(chan string, 1024)
	errCh := make(chan error, 1)
	counter := make(chan struct{}, 1024)
	resChan := make(chan string, 1024)
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(1)
	// 启动一个goroutine向taskCh中投递任务
	go func() {
		defer wg.Done()
		for _, file := range goFiles {
			taskCh <- file
		}
	}()
	// 启动8个goroutine来处理文件
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				// 等待退出信号
				case <- ctx.Done():
					return
				// 从taskCh中获取任务
				case path := <-taskCh:
					// 最多读取2K
					content, err := readFile(path, 2048)
					if err != nil {
						color.Red("read file %v error, %v", path, err)
						errCh <- err
						return
					}
					// 判断文件中是否有//go:generate
					if bytes.Contains(content, []byte("//go:generate")) {
						resChan <- path
					}

					// 通知主goroutine完成一个任务
					counter <- struct{}{}
				}
			}

		}()
	}

	n := 0
loop:
	for {
		select {
		// 发生错误，结束后续任务
		case <- errCh:
			cancel()
			return err
		// 统计任务完成情况，一旦任务完成就通知其它goroutine退出
		case <-counter:
			n++
			if n == len(goFiles) {
				cancel()
				break loop
			}
		// 获取结果
		case f := <-resChan:
			args = append(args, f)
		}
	}

	// 等待其它goroutine退出
	wg.Wait()

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