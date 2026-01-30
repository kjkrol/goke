package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

type ViewData struct {
	N             int
	AllTParams    string
	Indices       []int
	HIndices      []int
	HTParams      string
	HasTail       bool
	TIndices      []int
	TTParams      string
	PHIndices     []int
	PHTParams     string
	HasVTail      bool
	PTIndices     []int
	PTTParams     string
	AllSeqType    string
	AllYieldArgs  string
	PureSeqType   string
	PureYieldArgs string
}

func main() {
	// 1. Find template file
	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)
	tmplPath := filepath.Join(baseDir, "view.go.tmpl")

	// 2. Prepare template helper functions
	funcMap := template.FuncMap{
		"additions": func(a, b int) int { return a + b },
	}

	// 3. Load template
	tmpl, err := template.New("view.go.tmpl").Funcs(funcMap).ParseFiles(tmplPath)
	if err != nil {
		fmt.Printf("Template file not found w: %s\n", tmplPath)
		panic(err)
	}

	// 4. Generate view fies
	for i := 1; i <= 8; i++ {
		data := prepareData(i)
		fileName := fmt.Sprintf("view_gen_%d.go", i)

		f, err := os.Create(fileName)
		if err != nil {
			fmt.Printf("Can not create file %s: %v\n", fileName, err)
			continue
		}

		err = tmpl.Execute(f, data)
		f.Close()

		if err != nil {
			fmt.Printf("Error during genartion %s: %v\n", fileName, err)
			continue
		}

		// 5. Code formating
		exec.Command("go", "fmt", fileName).Run()
	}

	fmt.Println("SUccess: 8 view files has been generated.")
}

func prepareData(n int) ViewData {
	allTypes := make([]string, n)
	indices := make([]int, n)
	for i := 0; i < n; i++ {
		allTypes[i] = fmt.Sprintf("T%d", i+1)
		indices[i] = i
	}

	hCount := n
	if hCount > 3 {
		hCount = 3
	}
	phCount := n
	if phCount > 4 {
		phCount = 4
	}

	data := ViewData{
		N:          n,
		AllTParams: strings.Join(allTypes, ", "),
		Indices:    indices,
		HIndices:   indices[:hCount],
		HTParams:   strings.Join(allTypes[:hCount], ", "),
		PHIndices:  indices[:phCount],
		PHTParams:  strings.Join(allTypes[:phCount], ", "),
	}

	// Configure All/Filter
	if n > hCount {
		data.HasTail = true
		data.TIndices = indices[hCount:]
		data.TTParams = strings.Join(allTypes[hCount:], ", ")
		data.AllSeqType = fmt.Sprintf("iter.Seq2[Head%d[%s], Tail%d[%s]]", n, data.HTParams, n, data.TTParams)
		data.AllYieldArgs = fmt.Sprintf("Head%d[%s], Tail%d[%s]", n, data.HTParams, n, data.TTParams)
	} else {
		data.AllSeqType = fmt.Sprintf("iter.Seq[Head%d[%s]]", n, data.HTParams)
		data.AllYieldArgs = fmt.Sprintf("Head%d[%s]", n, data.HTParams)
	}

	// Configure Values/FilterValues
	if n > phCount {
		data.HasVTail = true
		data.PTIndices = indices[phCount:]
		data.PTTParams = strings.Join(allTypes[phCount:], ", ")
		data.PureSeqType = fmt.Sprintf("iter.Seq2[VHead%d[%s], VTail%d[%s]]", n, data.PHTParams, n, data.PTTParams)
		data.PureYieldArgs = fmt.Sprintf("VHead%d[%s], VTail%d[%s]", n, data.PHTParams, n, data.PTTParams)
	} else {
		data.PureSeqType = fmt.Sprintf("iter.Seq[VHead%d[%s]]", n, data.PHTParams)
		data.PureYieldArgs = fmt.Sprintf("VHead%d[%s]", n, data.PHTParams)
	}

	return data
}
