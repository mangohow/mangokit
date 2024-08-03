package projectcmd

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var CmdProject = &cobra.Command{
	Use:     "create",
	Short:   "Create a new project based on the preset project",
	Long:    "Create a new project based on the preset project.",
	Example: "mangokit create helloworld github.com/xxx/helloworld",
	Run: func(cmd *cobra.Command, args []string) {
		projectDir := ""
		projectName := ""
		if len(args) < 2 {
			prompt1 := &survey.Input{
				Message: "What is project dir name",
				Help:    "Create project dir name",
			}

			err := survey.AskOne(prompt1, &projectDir)
			if err != nil || projectDir == "" {
				return
			}

			prompt2 := &survey.Input{
				Message: "What is project name or go mod name",
				Help:    "Input go mod name",
			}

			err = survey.AskOne(prompt2, &projectName)
			if err != nil || projectName == "" {
				return
			}
		} else {
			projectDir = args[0]
			projectName = args[1]
		}

		// 获取工作目录
		dir, err := os.Getwd()
		if err != nil {
			color.Red("get current work dir failed: %v", err)
			return
		}

		project := ProjectGenerator{
			DirName:     projectDir,
			ProjectName: projectName,
			BaseDir:     filepath.Join(dir, projectDir),
			RepoUrl:     repoUrl,
			Branch:      branch,
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
		defer cancel()
		if err = project.Generate(ctx); err != nil {
			color.Red("generate project failed")
		}
	},
}

var (
	repoUrl = "https://github.com/mangohow/mangokit-template.git"
	branch  = ""
	timeout = "60s"

	timeoutDuration = time.Second * 60
)

func init() {
	CmdProject.Flags().StringVarP(&repoUrl, "repo-url", "r", repoUrl, "template repo")
	CmdProject.Flags().StringVarP(&branch, "branch", "b", branch, "template repo branch")
	CmdProject.Flags().StringVarP(&timeout, "timeout", "t", timeout, "pull project timeout")

	duration, err := time.ParseDuration(timeout)
	if err == nil {
		timeoutDuration = duration
	}
}
