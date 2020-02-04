package files

import (
	"fmt"
	"log"
	"strings"
	"io/ioutil"
	"path/filepath"
	"pbg/graph"
)

// Add a source file to a PBG
func addSourceFile(pbg *graph.ProgramBehaviorGraph, sourceFile string) {
	sourceDir, sourceFile := filepath.Split(sourceFile);

	log.Printf("Adding source file %s from %s\n", sourceFile, sourceDir	);

	pbg.AddRelation(sourceDir, "contains-file", sourceFile);

	data, err := ioutil.ReadFile(sourceDir + sourceFile);

	if err != nil {
		panic(err);
	}

	content := string(data)
	pbg.AddRelation(sourceFile, "has-text", content);

	lines := strings.Split(content, "\n")	
	index := 1

	pbg.AddRelationFunc(func(ch chan []string) {
		for _, line := range lines {
			lineStr := fmt.Sprintf("line-%d", index)

			ch <- []string { sourceFile, "has-line", lineStr }
			ch <- []string { sourceFile + ":"+ lineStr, "line-content", line }

			index += 1
		}

		close(ch);
	});

	log.Printf("Added %d lines of code\n", index)
}

// Add all files found to PBG
func addSourceFiles(pbg *graph.ProgramBehaviorGraph, sourceFiles []string) {
	for _, sourceFile := range sourceFiles {
		// Make sure we didn't get a directory
		if filepath.Base(sourceFile) == "" {
			panic("directory found in glob");
		}

		addSourceFile(pbg, sourceFile);
	}
}

func loadsFiles(pbg *graph.ProgramBehaviorGraph, opt map[string] interface{}) {
	log.Printf("Loading stuff: %v\n", opt);
	// First attempt to load from a glob
	sourceGlob, ok := opt["sourceGlob"].(string);

	if ok  {
		fmt.Printf("Loading glob %s\n", sourceGlob)

		sourceFiles, err := filepath.Glob(sourceGlob);

		if err != nil { 
			panic(err);
		}

		fmt.Printf("Files found: %v\n", sourceFiles);

		addSourceFiles(pbg, sourceFiles);
	} else {
		fmt.Println("No glob found")
	}

	// Then attempt to load directly referenced files
	_sourceFiles, ok := opt["sourceFiles"].([]interface {});

	if ok {
		sourceFiles := make([]string, 0)

		for _, sourceFile := range _sourceFiles {
			sourceFiles = append(sourceFiles, sourceFile.(string))
		}

		addSourceFiles(pbg, sourceFiles);
	}
}

func init() {
	graph.RegisterProvider("files", loadsFiles);
}