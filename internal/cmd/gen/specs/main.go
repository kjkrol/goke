package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
		"typeList": func(n int) string {
			if n == 0 {
				return ""
			}
			parts := make([]string, n)
			for i := 0; i < n; i++ {
				parts[i] = fmt.Sprintf("T%d any", i+1)
			}
			return "[" + strings.Join(parts, ", ") + "]"
		},
		"typeArgs": func(n int) string {
			if n == 0 {
				return ""
			}
			parts := make([]string, n)
			for i := 0; i < n; i++ {
				parts[i] = fmt.Sprintf("T%d", i+1)
			}
			return "[" + strings.Join(parts, ", ") + "]"
		},
	}

	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)
	tmplPath := filepath.Join(baseDir, "spec.go.tmpl")
	tmpl, err := template.New("spec.go.tmpl").Funcs(funcMap).ParseFiles(tmplPath)
	if err != nil {
		fmt.Printf("Template file not found at: %s\n", tmplPath)
		panic(err)
	}

	f, _ := os.Create("internal/comp/spec_gen.go")
	defer f.Close()

	tmpl.Execute(f, map[string]any{
		"Range": []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	})
}
