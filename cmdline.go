package main

import (
	"fmt"
	"os"
	// "log"
	_ "pbg/files"
	_ "pbg/clang"
	_ "pbg/elf"
	_ "pbg/trace"
	_ "pbg/disasm"
)

func genUsage() {
	fmt.Printf("Usage: %s [cmd ...]\n", os.Args[0])
	fmt.Println("List of available commands:")
	fmt.Println("\tdatabase")
	fmt.Println("\tproject")
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		genUsage()
	}

	// f, err := os.OpenFile("pbg.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	// if err != nil {
	// 	log.Fatalf("error opening file: %v", err)
	// } else {
	// 	defer f.Close()
	// 	log.SetOutput(f)
	// }

	switch os.Args[1] {
	case "database": databaseCmd();
	case "project": projectCmd();
	default:
		fmt.Printf("Unknown command %s\n", os.Args[1])
		genUsage()
	}
}
