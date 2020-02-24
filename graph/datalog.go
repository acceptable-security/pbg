package graph;

import (
	"fmt"
	"bufio"
	"os"
)

func (pbg *ProgramBehaviorGraph) GenerateDatalog(filename string) {
	foundRels := make(map [string] bool)

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)

	if err != nil {
		panic(err)
	}

	writer := bufio.NewWriter(file)

	ch := pbg.QueryTripletAsync("g.V().Tag('subject').Out(null, 'predicate').Tag('object').All()")

	for triplet := range ch {
		if _, ok := foundRels[triplet.predicate]; !ok {
			foundRels[triplet.predicate] = true;
			writer.WriteString(fmt.Sprintf(".decl %s(from symbol, to symbol)\n", triplet.predicate))
		}

		writer.WriteString(fmt.Sprintf("%s(\"%s\", \"%s\")\n", triplet.predicate, triplet.subject, triplet.object))
	}

	writer.Flush()
	file.Close()
}