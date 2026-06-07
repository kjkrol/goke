package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"text/template"
)

func main() {
	funcMap := template.FuncMap{
		"seq": func(start, end int) []int {
			var res []int
			for i := start; i <= end; i++ {
				res = append(res, i)
			}
			return res
		},
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
	}

	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)
	tmplPath := filepath.Join(baseDir, "view.go.tmpl")
	tmpl, err := template.New("view.go.tmpl").Funcs(funcMap).ParseFiles(tmplPath)
	if err != nil {
		fmt.Printf("Template file not found at: %s\n", tmplPath)
		panic(err)
	}

	f, _ := os.Create("views_gen.go")
	defer f.Close()

	tmpl.Execute(f, map[string]any{
		"Range": []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	})
}
