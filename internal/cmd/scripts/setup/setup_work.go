package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	if _, err := os.Stat("go.work"); os.IsNotExist(err) {
		_ = exec.Command("go", "work", "init").Run()
	}

	args := []string{"work", "use", "."}

	err := filepath.Walk("examples", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Base(path) == "go.mod" {
			args = append(args, filepath.Dir(path))
		}
		return nil
	})

	if err != nil {
		os.Exit(1)
	}

	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}
