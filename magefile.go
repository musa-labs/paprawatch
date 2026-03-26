//go:build mage
package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/magefile/mage/mg"
)

// Default target to run when none is specified
var Default = Install

// InstallDeps installs all project dependencies.
func InstallDeps() error {
	fmt.Println("Installing dependencies...")
	return run("go", "mod", "download")
}

// Install builds and installs the paprawatch binary to $GOPATH/bin.
func Install() error {
	mg.Deps(InstallDeps)
	fmt.Println("Installing paprawatch...")
	return run("go", "install", ".")
}

// Test runs all tests in the project.
func Test() error {
	mg.Deps(InstallDeps)
	fmt.Println("Running tests...")
	return run("go", "test", "-v", "./...")
}

// Clean removes any built binaries.
func Clean() error {
	fmt.Println("Cleaning up...")
	return os.Remove("paprawatch")
}

// Helper function to run shell commands
func run(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
