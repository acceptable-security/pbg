package trace;

import (
	"log"
	"io/ioutil"
	"os"
	"os/exec"
	"bufio"
	"pbg/graph"
	"strings"
	"path"
)

func loadRawInstrTrace(pbg *graph.ProgramBehaviorGraph, opt map[string] interface{}) {
	cmdLine, ok := opt["cmdLine"].(string)

	if !ok {
		log.Printf("No instruction trace command found, skipping...\n")
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

	args := []string{ "-c", "instrace", "-verbose", "5", "--" }

	for _, arg := range strings.Split(cmdLine, " ") {
		args = append(args, arg)
	}

	log.Printf("Executing instrace (%s %s)...\n", total_path, args)

	cmdObj := exec.Command(total_path, args...)


	stdoutObj, err := cmdObj.StderrPipe()

	if err != nil {
		panic(err)
		return
	}


	cmdObj.Start()

	slurp, _ := ioutil.ReadAll(stdoutObj)
	log.Printf("Output: %s\n", slurp)

	log_path := ""

	for _, line := range strings.Split(string(slurp), "\n") {
		if strings.Contains(line, "Data file ") {
			line = line[len("Data file "):]
			line = line[:len(line)-len(" created")]

			log_path = line
			break
		}
	}

	if log_path == "" {
		panic("Failed to find log path in rawinstrace result")
	}

	file, err := os.Open(log_path)

	if err != nil {
		panic(err)
	}

	defer file.Close()

	count := 0

	pbg.AddRelationFunc(func(ch chan []string) {
		idx := 0
		var last string

		scanner := bufio.NewScanner(file)
		
		for scanner.Scan() {
			text := scanner.Text()

			if idx == 0 || len(text) == 0 {
				idx += 1
				continue
			}

			idx += 1;

			parts := strings.Split(text, ",")

			if len(parts) != 2 {
				continue
			}

			if ( idx > 1 ) {
				count += 1
				ch <- []string{ last, "next-address", parts[0] }
			}

			last = parts[0]
		}

		if err := scanner.Err(); err != nil {
			panic(err)
		}

		close(ch)
	})

	log.Printf("Loaded %d instructions in trace\n", count)
}

func init() {
	graph.RegisterProvider("rawinstrtrace", loadRawInstrTrace, "elf")
}