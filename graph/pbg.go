package graph

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	_ "github.com/cayleygraph/cayley/graph/all"
	"github.com/cayleygraph/cayley/quad"
	"github.com/cayleygraph/cayley/query"
	_ "github.com/cayleygraph/cayley/query/gizmo"
	"github.com/cayleygraph/cayley/query/gizmo"
)

const PBG_QUERY_LIMIT = 2500;

type ProgramBehaviorGraph struct {
	store *cayley.Handle	
	session query.Session
	options map [string] map[string] interface{}

	// Automatically handle bulking
	autoBulk int
	bulkBuf [][]string

	reservoirIndex int
	reservoirSize int
	reservoir [][]string 
}

// Constructs a new ProgramBehaviorGraph object from a dbpath and handler.
// For now it is assumed that a Gizmo query tool is desired, however, this
// might change in the future.
func NewPBG(db string, path string, init bool) (*ProgramBehaviorGraph, error) {
	var store *cayley.Handle
	var err error

	if init {
		err := graph.InitQuadStore(db, path, nil)

		if err != nil {
			return nil, err;
		}

		store, err = cayley.NewGraph(db, path, nil)

		if err != nil {
			return nil, err;
		}
	} else {
		store, err = cayley.NewGraph(db, path, nil)

		if err != nil {
			return nil, err;
		}
	}

	obj := new(ProgramBehaviorGraph)

	obj.store = store;
	obj.options = make(map [string] map[string] interface{})

	// Temporarily disabled while graphql doesn't build
	obj.session = query.NewSession(store, "gizmo")

	return obj, nil;
}

// Sets the options for a given provider
func (pbg *ProgramBehaviorGraph) SetOptions(provider string, options map[string] interface{}) {
	pbg.options[provider] = options
}

// Adds a triplet relation (subject, verb, object) to the graph. This API may change
// in the future to support contexts.
func (pbg *ProgramBehaviorGraph) AddRelation(from string, rel string, to string) {
	if pbg.autoBulk > 0 {
		pbg.bulkBuf = append(pbg.bulkBuf, []string{ from, rel, to });

		if len(pbg.bulkBuf) > pbg.autoBulk {
			pbg.AddRelationBulk(pbg.bulkBuf)
			pbg.bulkBuf = pbg.bulkBuf[:0]
		}
	} else {
		pbg.store.AddQuad(quad.Make(from, rel, to, nil));
	}
}

// Enable functionality for AddRelation to use in memory-bulking
func (pbg *ProgramBehaviorGraph) SetAutoBulk(count int) {
	pbg.autoBulk = count

	if count > 0 {
		pbg.bulkBuf = make([][]string, 0)
	} else {
		// Left over in bulk buffer
		if len(pbg.bulkBuf) > 0 {
			pbg.AddRelationBulk(pbg.bulkBuf)
			pbg.bulkBuf = pbg.bulkBuf[:0]
		}
	}
}

// Enable reservoir handling
func (pbg *ProgramBehaviorGraph) SetReservoir(count int) {
	rand.Seed(time.Now().UnixNano())

	pbg.reservoirIndex = 0
	pbg.reservoirSize = count

	if count > 0 {
		pbg.reservoir = make([][]string, 0)
	} else {
		if len(pbg.reservoir) > 0 {
			// Commit reservoir
			pbg.AddRelationBulk(pbg.reservoir)
			pbg.reservoir = pbg.reservoir[:0]
		}
	}
}

// Add to reservoir
func (pbg *ProgramBehaviorGraph) AddRelationReservoir(data [][]string) {
	for _, piece := range data {
		if pbg.reservoirIndex < pbg.reservoirSize {
			pbg.reservoir = append(pbg.reservoir, piece)
		} else {
			if rnd := rand.Intn(pbg.reservoirIndex); rnd < pbg.reservoirSize {
				pbg.reservoir[rnd] = piece
			}
		}
	}
}

// Executes a bulk form of AddRelation. Assumes that the array contains a list of 3-lists
func (pbg *ProgramBehaviorGraph) AddRelationBulk(data [][]string) {
	if pbg.reservoirSize > 0 {
		pbg.AddRelationReservoir(data)
		return
	}

	quads := make([]quad.Quad, len(data));

	for _, piece := range data {
		if len(piece) != 3 {
			continue;
		}

		quads = append(quads, quad.Make(piece[0], piece[1], piece[2], ""));
	}

	writer := graph.NewWriter(pbg.store.QuadWriter)
	_, err := writer.WriteQuads(quads)

	if err != nil {
		panic(err);
	}
}

// Executes a function and chunks up its output to add to addRealtionBulk
func (pbg *ProgramBehaviorGraph) AddRelationFunc(produce func (chan []string)) {
	relationChannel := make(chan []string, 0)

	go produce(relationChannel)

	tmpBuffer := make([][]string, 0);
	upperBound := 10000

	for output := range relationChannel {
		tmpBuffer = append(tmpBuffer, output)

		if len(tmpBuffer) >= upperBound {
			pbg.AddRelationBulk(tmpBuffer);
			tmpBuffer = tmpBuffer[:0];
			upperBound = upperBound * 2
		}
	}

	if len(tmpBuffer) > 0 {
		pbg.AddRelationBulk(tmpBuffer);
	}
}


// Execute a query based on the loaded session and return results/errors.
func (pbg *ProgramBehaviorGraph) Query(qu string) ([]string, error) {
	ctx := context.TODO()

	it, err := pbg.session.Execute(ctx,  qu, query.Options{
			Collation: query.Raw,
			Limit: PBG_QUERY_LIMIT,
	})

	if err != nil {
		panic(err)
	}

	var results []string

	for it.Next(ctx) {
		data := it.Result().(*gizmo.Result)

		if data.Val == nil {
			if val := data.Tags[gizmo.TopResultTag]; val != nil {
				found_str := quad.StringOf(pbg.store.NameOf(val));
				results = append(results, found_str);
			} else {
				panic(fmt.Sprintf("Unknown data %v", data))
			}
		} else {
			switch val := data.Val.(type) {
			case string:
				results = append(results, val);
			default:
				results = append(results, fmt.Sprint(val));
			}
		}
	}

	if err := it.Err(); err != nil {
		panic(err)
		return nil, err;
	}

	return results, nil;
}

type PBGTriplet struct {
	subject string
	predicate string
	object string
}

func (pbg *ProgramBehaviorGraph) QueryTriplet(qu string) ([]PBGTriplet, error) {
	ctx := context.TODO()

	it, err := pbg.session.Execute(ctx,  qu, query.Options{
			Collation: query.Raw,
			Limit: PBG_QUERY_LIMIT,
	})

	if err != nil {
		panic(err)
	}

	var results []PBGTriplet

	for it.Next(ctx) {
		data := it.Result().(*gizmo.Result)

		if data.Val == nil {
			subject := quad.StringOf(pbg.store.NameOf(data.Tags["subject"]))
			predicate := quad.StringOf(pbg.store.NameOf(data.Tags["predicate"]))
			object := quad.StringOf(pbg.store.NameOf(data.Tags["object"]))

			results = append(results, PBGTriplet { subject, predicate, object })
		} else {
			panic("Failed to get triplet")
		}
	}

	if err := it.Err(); err != nil {
		panic(err)
		return nil, err;
	}

	return results, nil;
}