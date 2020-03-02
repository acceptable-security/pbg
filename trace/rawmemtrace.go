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

	log.Printf("Executing drcachesim (%s %s)...\n", total_path, args)


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


	// Input parsing similar to memtrace.py.
	pbg.AddRelationFunc(func(ch chan []string) {
		scanner := bufio.NewScanner(stderrObj)
		
		for scanner.Scan() {
			text := scanner.Text()

			if len(text) < 3 || text[:2] != "::" || !strings.Contains(text, "@") {
				continue
			}

			cmd := strings.Split(text[strings.Index(text, "@") + 1:], " ");

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
 
	slurp, _ := ioutil.ReadAll(stdoutObj)
	log.Printf("Output: %s\n", slurp)

	file, err := os.Open("./.tmp_cache.gz")

	if err != nil {
		panic(err)
	}

	defer file.Close()

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