package graph;

import (
	"bufio"
	"os"
)

func (pbg *ProgramBehaviorGraph) GenerateDatalog(query string, filename string) {
	var foundRels map [string] bool

	file, err := os.Open(filename)

	if err != nil {
		panic(err)
	}

	defer file.Close()
	writer := bufio.NewWriter(file)

	ch := pbg.QueryTripletAsync("g.V().Tag('subject').Out(null, 'predicate').Tag('object').All()")

	for triplet := range ch {
		if _, ok := foundRels[triplet.predicate]; !ok {
			foundRels[triplet.predicate] = true;
			writer.WriteString(".decl " + triplet.predicate + "(from symbol, to symbol)\n")
		}

		writer.WriteString(triplet.predicate + "(\"" + triplet.subject + ", " + triplet.object + ")")
	}
}