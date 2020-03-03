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
	ch := pbg.QueryTripletAsync("g.V().Tag('subject').Out(null, 'predicate').Tag('object').All()")

	files := make(map[string] DestFile)

	for triplet := range ch {
		pred := strings.ReplaceAll(triplet.predicate[1:len(triplet.predicate)-1], "-", "_")
		writer := getFilePred(filename, pred, files)
		subj := triplet.subject[1:len(triplet.subject)-1]
		obj := triplet.object[1:len(triplet.object)-1]
		writer.WriteString(fmt.Sprintf("%s\t%s\n", subj, obj))
	}

	for _, obj := range files {
		obj.Writer.Flush()
		obj.File.Close()
	}
}