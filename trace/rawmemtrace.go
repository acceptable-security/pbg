package trace;

import (
	"compress/gzip"
	"log"
	"os"
	"os/exec"
	"bufio"
	"pbg/graph"
	"strings"
	"path"
)

func loadRawMemTrace(pbg *graph.ProgramBehaviorGraph, opt map[string] interface{}) {
	cmdLine, ok := opt["cmdLine"].(string)

	if !ok {
		log.Printf("No memory trace command found, skipping...\n")
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

	args := []string{ "-t", "drcachesim", "-verbose", "6", "-LL_miss_file", "./.tmp_cache.gz", "--" }

	for _, arg := range strings.Split(cmdLine, " ") {
		args = append(args, arg)
	}

	cmdObj := exec.Command(total_path, args...)

	log.Printf("Executing drcachesim...\n")

	// Start execution and get stdout
	cmdObj.Start()
	stdoutObj, err := cmdObj.StdoutPipe()

	if err != nil {
		panic(err)
		return
	}

	// Input parsing similar to memtrace.py.
	pbg.AddRelationFunc(func(ch chan []string) {
		scanner := bufio.NewScanner(stdoutObj)
		
		for scanner.Scan() {
			text := scanner.Text()

			if text[:2] != "::" || !strings.Contains(text, "@") {
				continue
			}

			cmd := strings.Split(text[strings.Index(text, "@"):], " ");

			if len(cmd) < 2 || (cmd[1] != "read" && cmd[1] != "write") {
				continue
			}


			ch <- []string{ cmd[0], cmd[1] + "-address", cmd[2] }
		}

		if err := scanner.Err(); err != nil {
			panic(err)
		}

		close(ch)
	})

	file, err := os.Open("./.tmp_cache.gz")

	if err != nil {
		panic(err)
	}

	reader, err := gzip.NewReader(file)

	if err != nil {
		panic(err)
	}

	pbg.AddRelationFunc(func(ch chan []string) {
		scanner := bufio.NewScanner(reader)

		for scanner.Scan() {
			elements := strings.Split(scanner.Text(), ",")

			// Skip malformed entries
			if len(elements) != 2 {
				continue
			}

			ch <- []string{ elements[0], "miss-address", elements[1] }
		}

		if err := scanner.Err(); err != nil {
			panic(err)
		}

		close(ch)
	})
}

func init() {
	graph.RegisterProvider("rawmemtrace", loadRawMemTrace, "elf")
}