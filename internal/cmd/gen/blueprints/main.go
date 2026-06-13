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
			parts := make([]string, n)
			for i := 0; i < n; i++ {
				parts[i] = fmt.Sprintf("T%d any", i+1)
			}
			return strings.Join(parts, ", ")
		},
		"typeArgs": func(n int) string {
			parts := make([]string, n)
			for i := 0; i < n; i++ {
				parts[i] = fmt.Sprintf("T%d", i+1)
			}
			return strings.Join(parts, ", ")
		},
	}

	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)

	data := map[string]any{
		"Range": []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}

	generate := func(tmplName, outPath string) {
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

	generate("ent_factory.go.tmpl", "internal/ent/factories_gen.go")
	generate("goke_blueprint.go.tmpl", "aliases_factories_gen.go")
}
