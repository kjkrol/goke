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
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
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
		"compNames": func(n int) string {
			if n == 0 {
				return "no specific components (filter-only)"
			}
			parts := make([]string, n)
			for i := 0; i < n; i++ {
				parts[i] = fmt.Sprintf("T%d", i+1)
			}
			return strings.Join(parts, ", ")
		},
	}

	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)

	generate := func(tmplName, outPath string, data any) {
		tmplPath := filepath.Join(baseDir, tmplName)
		tmpl, err := template.New(tmplName).Funcs(funcMap).ParseFiles(tmplPath)
		if err != nil {
			fmt.Printf("Template file not found at: %s\n", tmplPath)
			panic(err)
		}
		f, err := os.Create(outPath)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		tmpl.Execute(f, data)
	}

	generate("query_view.go.tmpl", "internal/query/views_gen.go",
		map[string]any{"Range": []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}})
	generate("goke_view.go.tmpl", "aliases_views_gen.go",
		map[string]any{"Range": []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}})
}
