package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func main() {
	count := "20"
	if len(os.Args) > 1 {
		count = os.Args[1]
	}

	fmt.Printf("--- Performance Analysis | Iterations: %s ---\n", count)

	// TARGET: Dokładna ścieżka do internal/bench bez rekurencyjnego "..."
	// GOWORK=off: Kluczowe, by olać ciężkie moduły z examples/
	cmd := exec.Command("go", "test",
		"-bench=.",
		"-benchmem",
		"-count="+count,
		"-p=1",
		"-cpu=1",
		"../../../bench",
	)

	cmd.Env = append(os.Environ(), "GOWORK=off")

	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Printf("Fatal: Command start failed: %v\n", err)
		return
	}

	results := make(map[string][]float64)
	memResults := make(map[string][]float64)
	re := regexp.MustCompile(`(Benchmark[^\s-]+)\s+\d+\s+([\d.]+)\s+ns/op\s+([\d.]+)\s+B/op`)

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line) // Widzisz progres na żywo

		if m := re.FindStringSubmatch(line); m != nil {
			ns, _ := strconv.ParseFloat(m[2], 64)
			mem, _ := strconv.ParseFloat(m[3], 64)
			results[m[1]] = append(results[m[1]], ns)
			memResults[m[1]] = append(memResults[m[1]], mem)
		}
	}

	cmd.Wait()
	printFinalTable(results, memResults)
}

func printFinalTable(res map[string][]float64, mem map[string][]float64) {
	if len(res) == 0 {
		fmt.Println("\nNo benchmarks found in internal/bench.")
		return
	}
	fmt.Printf("\n%-45s | %-12s | %-12s | %-12s\n", "Benchmark", "Avg (ns)", "P95 (ns)", "Mem B/op")
	fmt.Println(strings.Repeat("-", 85))

	keys := make([]string, 0, len(res))
	for k := range res {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		values := res[name]
		sort.Float64s(values)

		var sumTime, sumMem float64
		for i, v := range values {
			sumTime += v
			sumMem += mem[name][i]
		}

		p95Idx := int(float64(len(values)) * 0.95)
		if p95Idx >= len(values) {
			p95Idx = len(values) - 1
		}

		fmt.Printf("%-45s | %-12.2f | %-12.2f | %-12.0f\n",
			name, sumTime/float64(len(values)), values[p95Idx], sumMem/float64(len(values)))
	}
}
