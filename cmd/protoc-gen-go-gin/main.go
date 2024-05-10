package main

import (
	"flag"
	"fmt"
	"os"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

var (
	showVersion = flag.Bool("version", false, "print the version and exit")
)

var (
	debug = true
	output *os.File
)

func init() {
	if !debug {
		return
	}
	var err error
	output, err = os.OpenFile("protoc-gen-go-gin.log", os.O_CREATE | os.O_RDWR, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open log file error: %v", err)
		os.Exit(1)
	}
}

func main() {
	if debug {
		defer output.Close()
	}

	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-go-gin %v\n", version)
		return
	}

	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(plugin *protogen.Plugin) error {
		plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range plugin.Files {
			if !f.Generate {
				continue
			}

			generateFile(plugin, f)
		}

		return nil
	})
}
