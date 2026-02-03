package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// 1. Inicjalizacja go.work jeśli nie istnieje
	if _, err := os.Stat("go.work"); os.IsNotExist(err) {
		_ = exec.Command("go", "work", "init").Run()
	}
	
	// 2. Szukanie wszystkich modułów
	// Zaczynamy od roota (.)
	args := []string{"work", "use", "."}
	
	// Przeszukujemy katalog examples
	err := filepath.Walk("examples", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Jeśli znajdziemy go.mod, dodajemy folder do listy
		if !info.IsDir() && filepath.Base(path) == "go.mod" {
			args = append(args, filepath.Dir(path))
		}
		return nil
	})

	if err != nil {
		os.Exit(1)
	}
	
	// 3. Wykonanie go work use [moduły...]
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}

