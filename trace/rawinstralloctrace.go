package trace;

import (
	"compress/gzip"
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
		var last string

		for scanner.Scan() {
			text := scanner.Text()

			if len(text) >= 4 && text[:4] == "free" {

			} else if len(text) >= 6 && text[:6] == "malloc" {

			} else if len(text) >= 7 && text[:7] == "realloc" {

			} else {
				if idx == 0 || len(text) == 0 {
					idx += 1
					continue
				}

				idx += 1

				if ( idx > 1 ) {
					// Form step nodes
					lastIndex := fmt.Sprintf("step-%d", count)
					count += 1
					index := fmt.Sprintf("step-%d", count)

					// Add tuples into database
					ch <- []string{ last, "next-address", text }
					ch <- []string{ index, "step-address", last }
					ch <- []string { lastIndex, "next-step", index }
				}

				last = parts[0]
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