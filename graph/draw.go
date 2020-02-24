package graph;

import (
	"github.com/emicklei/dot"
	"os"
	"bufio"
)

func (pbg *ProgramBehaviorGraph) Draw(query string, filename string) {
	out, err := pbg.QueryTriplet(query)

	if err != nil {
		panic(err)
	}

	file, err := os.Open(filename)

	if err != nil {
		panic(err)
	}

	defer file.Close()
	writer := bufio.NewWriter(file)

	g := dot.NewGraph(dot.Directed)
	g.Attr("size", "7.75,10.25")
	g.Attr("ratio", "compress")
	nodes := make(map[string] dot.Node, 0)

	for _, item := range out {
		if _, ok := nodes[item.subject]; !ok {
			nodes[item.subject] = g.Node(item.subject)
		}

		if _, ok := nodes[item.object]; !ok {
			nodes[item.object] = g.Node(item.object)
		}

		nodes[item.subject].Edge(nodes[item.object], item.predicate)
	}

	writer.WriteString(g.String())
}