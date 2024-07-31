package main

import (
	"fmt"
	"testing"
)

func TestValidatePath(t *testing.T) {
	paths := []string{
		"/api/users",
		"/api/user/:id",
		"/api/user/:username/:password",
		"/api/user/",
		"api/user/:id",
		"/api/user/id/",
		"/api/:name/id",
		"/:name",
		"/api/:",
		"/:",
		"/",
		"/123",
		"/abc123",
		"/abc/123",
		"/123/abc",
		"/abc/:123",
		"/abc/:123abc",
		"/abc/:abc123",
		"",
		"/test/:a:b",
	}

	for _, path := range paths {
		fmt.Printf("Path: %s, Valid: %t\n", path, validatePath(path))
	}
}
