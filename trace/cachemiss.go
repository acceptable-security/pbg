package trace;

import (
	"log"
	"os"
	"bufio"
	"pbg/graph"
	"strings"
)

func loadCacheMiss(pbg *graph.ProgramBehaviorGraph, opt map[string] interface{}) {
	trace, ok := opt["cacheMissFile"].(string)

	if !ok {
		log.Println("No cache miss file found, skipping...")
		return;
	}

	file, err := os.Open(trace)

	if err != nil {
		panic(err);
	}

	defer file.Close()

	pbg.AddRelationFunc(func(ch chan []string) {
		idx := 0

		scanner := bufio.NewScanner(file)
		
		for scanner.Scan() {
			text := scanner.Text()

			if idx == 0 || len(text) == 0 {
				idx += 1;
				continue;
			}

			idx += 1;

			parts := strings.Split(text, ",")

			if len(parts) != 2 {
				continue
			}

			ch <- []string{ parts[0], "miss-address", parts[1] }
		}

		if err := scanner.Err(); err != nil {
			panic(err)
		}

		close(ch)
	})
}

func init() {
	graph.RegisterProvider("cachemiss", loadCacheMiss, "elf")
}