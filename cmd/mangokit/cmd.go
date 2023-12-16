package main

import (
	"github.com/mangohow/mangokit/cmd/mangokit/internal/addcmd"
	"github.com/mangohow/mangokit/cmd/mangokit/internal/generatecmd"
	"github.com/mangohow/mangokit/cmd/mangokit/internal/projectcmd"
	"github.com/spf13/cobra"
)

var rootCmd = cobra.Command{
	Use:     "mangokit",
	Short:   "mangokit is a toolkit for gin framework service",
	Long:    "mangokit is a toolkit for gin framework service, use proto to define service and error",
	Version: version,
}

func init() {
	rootCmd.AddCommand(projectcmd.CmdProject)
	rootCmd.AddCommand(generatecmd.CmdGenerate)
	rootCmd.AddCommand(addcmd.CmdAdd)
}
