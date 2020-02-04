package trace;

import (
	"log"
	"os"
	"bufio"
	"pbg/graph"
	"strings"
)

func loadInstrace(pbg *graph.ProgramBehaviorGraph, opt map[string] interface{}) {
	trace, ok := opt["traceFile"].(string)

	if !ok {
		log.Println("No instrace file found, skipping...");
		return;
	}

	file, err := os.Open(trace)

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
	graph.RegisterProvider("instrace", loadInstrace, "elf")
}