package graph;

import (
	"fmt"
	"bufio"
	"os"
	"strings"
)

func (pbg *ProgramBehaviorGraph) GenerateDatalog(filename string) {
	foundRels := make(map [string] bool)

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)

	if err != nil {
		panic(err)
	}

	defer file.Close()
	writer := bufio.NewWriter(file)

	ch := pbg.QueryTripletAsync("g.V().Tag('subject').Out(null, 'predicate').Tag('object').All()")

	for triplet := range ch {
		pred := strings.ReplaceAll(triplet.predicate[1:len(triplet.predicate)-1], "-", "_")

		if _, ok := foundRels[pred]; !ok {
			foundRels[pred] = true;
			writer.WriteString(fmt.Sprintf(".decl %s(from symbol, to symbol)\n", pred))
		}

		writer.WriteString(fmt.Sprintf("%s(%s, %s)\n", pred, triplet.subject, triplet.object))
	}

	writer.Flush()
}