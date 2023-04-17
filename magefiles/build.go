//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/target"
)

const appName = "costanza"

// Build builds the application binary
func Build() error {
	toUpdate, err := target.Dir(appName, "main.go", "cmd", "config", "internal")
	if err != nil {
		return fmt.Errorf("failed to get update deps: %w", err)
	}
	if toUpdate {
		fmt.Printf("Running build: go build -o %s .\n", appName)
		cmd := exec.Command("go", "build", "-o", appName, ".")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return nil
}

// Clean removes the application binary
func Clean() error {
	cmd := exec.Command("go", "clean")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
	return nil
}

// Rebuild rebuilds app from scratch
func Rebuild() error {
	mg.SerialDeps(Clean, Build)
	return nil
}

// Tests run all tests
func Tests() error {
	mg.Deps(Build)
	fmt.Println("running tests...")
	cmd := exec.Command("go", "test", "-v", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
