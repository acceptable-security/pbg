package trace;

import (
	"fmt"
	"log"
	"io/ioutil"
	"os"
	"os/exec"
	"bufio"
	"pbg/graph"
	"strings"
	"path"
)

func loadRawInstrAllocTrace(pbg *graph.ProgramBehaviorGraph, opt map[string] interface{}) {
	cmdLine, ok := opt["cmdLine"].(string)

	if !ok {
		log.Printf("No instralloc trace command found, skipping...\n")
		return
	}

	env := ""

	for _, text := range os.Environ() {
		tmp := strings.Split(text, "=")

		if len(tmp) != 2 || tmp[0] != "DYNAMORIO_HOME" {
			continue
		}

		env = tmp[1]
		break
	}

	if env == "" { 
		panic("Failed to find DYNAMORIO_HOME in path")
	}

	// Create execution object
	total_path := path.Join(env, "bin64/drrun")

	args := []string{ "-c", "./instr_alloc_client/build/libinstr_alloc.so", "--" }

	for _, arg := range strings.Split(cmdLine, " ") {
		args = append(args, arg)
	}

	log.Printf("Executing client (%s %s)...\n", total_path, args)

	cmdObj := exec.Command(total_path, args...)

	// Start execution and get stdout
	stderrObj, err := cmdObj.StderrPipe()

	if err != nil {
		panic(err)
		return
	}

	stdoutObj, err := cmdObj.StdoutPipe()

	if err != nil {
		panic(err)
		return
	}

	cmdObj.Start()

	pbg.AddRelationFunc(func(ch chan []string) {
		scanner := bufio.NewScanner(stderrObj)
		var lastPC string
		lastStep := "step-1"
		idx := 0
		count := 0

		for scanner.Scan() {
			text := scanner.Text()

			if len(text) >= 4 && text[:4] == "free" {
				var loc string

				fmt.Sscanf(text, "free %s", &loc)

				ch <- []string{ lastStep, "free-at", loc }
			} else if len(text) >= 6 && text[:6] == "malloc" {
				var amt, addr string

				fmt.Sscanf(text, "malloc %s %s", &amt, &addr)

				ch <- []string { lastStep, "malloc-amt", amt }
				ch <- []string { lastStep, "malloc-ptr", addr }
			} else if len(text) >= 7 && text[:7] == "realloc" {
				var amt, oldAddr, newAddr string

				fmt.Sscanf(text, "realloc %s %s %s", &oldAddr, &amt, &newAddr)

				ch <- []string { lastStep, "realloc-old-addr", oldAddr }
				ch <- []string { lastStep, "realloc-amt", amt }
				ch <- []string { lastStep, "realloc-new-addr", newAddr }
			} else if len(text) >= 6 && text[:6] == "calloc" {
				var amt, cnt, addr string

				fmt.Sscanf(text, "calloc %s %s %s", &amt, &cnt, &addr)

				ch <- []string { lastStep, "calloc-amt", amt }
				ch <- []string { lastStep, "calloc-cnt", cnt }
				ch <- []string { lastStep, "calloc-addr", addr }
			} else {
				if idx == 0 || len(text) == 0 {
					idx += 1
					continue
				}

				idx += 1

				if ( idx > 1 ) {
					// Form step nodes
					count += 1
					index := fmt.Sprintf("step-%d", count)

					// Add tuples into database
					ch <- []string{ lastPC, "next-address", text }
					ch <- []string{ index, "step-address", lastPC }
					ch <- []string{ lastStep, "next-step", index }

					lastStep = index
				}

				lastPC = text
			}
		}

		if err := scanner.Err(); err != nil {
			panic(err)
		}

		close(ch)
	})
 
	slurp, _ := ioutil.ReadAll(stdoutObj)
	log.Printf("Output: %s\n", slurp)
}

func init() {
	graph.RegisterProvider("rawinstralloctrace", loadRawInstrAllocTrace, "elf")
}