package internal

import (
	"os"
	"slices"
	"strings"
)

type Config struct {
	// 根据结构体名称或者tag来映射字段
	Mode string
	// 根据哪个tag来映射字段
	Tag string

	dir string
	// 浅拷贝还是深拷贝
	CopyMode string
}

type MappingKeyFunc func(name, tag string) string

// NameMappingKeyFunc 根据字段名称进行映射
func NameMappingKeyFunc(name, tag string) string {
	return name
}

// TagMappingKeyFunc 根据tag进行映射
func TagMappingKeyFunc(name, tag, targetTag string) string {
	tag = strings.Trim(tag, "`")
	tags := strings.Split(tag, " ")
	idx := slices.Index(tags, targetTag)
	if idx == -1 {
		return ""
	}
	// stmapper:"id"
	_, a, found := strings.Cut(tags[idx], ":")
	if !found {
		return ""
	}
	a = strings.Trim(a, "\"")
	b, _, found := strings.Cut(a, ",")
	if !found {
		return a
	}

	return b
}

type Generator struct {
	cfg Config

	parser *AstParser
}

func NewGenerator(cfg Config) *Generator {
	var mkf MappingKeyFunc
	if cfg.Mode == "name" {
		mkf = NameMappingKeyFunc
	} else if cfg.Mode == "tag" {
		mkf = func(name, tag string) string {
			return TagMappingKeyFunc(name, tag, cfg.Tag)
		}
	}

	return &Generator{
		cfg:    cfg,
		parser: NewAstParser(mkf),
	}
}

func (g *Generator) Execute() error {
	wd, err := os.Getwd()
	if err != nil {
		Fatalf("failed to get current working directory: %v", err)
	}

	pkgs, err := g.parser.ParseDir(wd)
	if err != nil {
		Fatalf("failed to parse directory: %v", err)
	}

	for _, pkg := range pkgs {
		if err := g.generatePackage(pkg); err != nil {
			Fatalf("failed to generate package: %v", err)
		}
	}

	return nil
}

func (g *Generator) generatePackage(pkg *Package) error {
	for _, file := range pkg.Files {
		if err := g.generateFile(file); err != nil {
			Errorf("failed to generate file: %v", err)
			return err
		}
	}

	return nil
}

func (g *Generator) generateFile(file *File) error {

	return nil
}
