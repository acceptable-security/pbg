package graph;

import (
	"fmt"
	"bufio"
	"os"
	"strings"
)

type DestFile struct {
	File *os.File
	Writer *bufio.Writer
};

func getFilePred(filename string, pred string, files map[string] DestFile) *bufio.Writer {
	if obj, ok := files[pred]; ok {
		return obj.Writer
	}

	file, err := os.OpenFile(filename + "/" + pred, os.O_RDWR|os.O_CREATE, 0755)

	if err != nil {
		panic(err)
	}

	writer := bufio.NewWriter(file)
	files[pred] = DestFile{ file, writer }

	return writer
}

func (pbg *ProgramBehaviorGraph) GenerateDatalog(filename string) {
	foundRels := make(map [string] bool)

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)

	if err != nil {
		panic(err)
	}

	defer file.Close()
	writer := bufio.NewWriter(file)

	ch := pbg.QueryTripletAsync("g.V().Tag('subject').Out(null, 'predicate').Tag('object').All()")

	files := make(map[string] DestFile)

	for triplet := range ch {
		pred := strings.ReplaceAll(triplet.predicate[1:len(triplet.predicate)-1], "-", "_")
		writer := getFilePred(filename, pred, files)

		if _, ok := foundRels[pred]; !ok {
			foundRels[pred] = true;
			writer.WriteString(fmt.Sprintf(".decl %s(from symbol, to symbol)\n", pred))
		}

		writer.WriteString(fmt.Sprintf("%s(%s, %s)\n", pred, triplet.subject, triplet.object))
	}

	for pred, obj := range files {
		obj.Writer.Flush()
		obj.File.Close()
	}
}