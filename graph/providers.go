package graph

import (
	"fmt"
	"log"
	"container/heap"
	"sync"
	"sort"
	"strings"
	"time"
)

type PBGProvider func (pbg *ProgramBehaviorGraph, opt map[string] interface{});

type PBGProviderDistCacheEntry struct {
	from string;
	to string
}

// Mutex locking all the following
var PBGProviderListMutex *sync.Mutex;

// Mapping between the provider name and its executing code
var PBGProviderList map[string] PBGProvider;

// A provider's dependencies. Root providers have no dependencies.
var PBGProviderBackDepList map[string] []string;

// List of providers that depend on a given string.
var PBGProviderForwardDepList map[string] []string;

// A cache for provider -> upstream dependency distances
var PBGProviderDistCache map[PBGProviderDistCacheEntry] int;

// Get the root providers with no dependencies
func getProviderRoots() []string {
	roots := make([]string, 0);

	PBGProviderListMutex.Lock();

	for provider, deps := range PBGProviderBackDepList {
		if len(deps) == 0 {
			roots = append(roots, provider);
		}
	}

	PBGProviderListMutex.Unlock();

	return roots;
}

// A dependency / root-distance pair
type PBGDepDist struct {
	dep string; // Dependency in question
	dist int;	// It's longest distance to the given dependency
}

// Initialize the global objects for provider lists
func initProviderObjects() {
	if PBGProviderList != nil {
		return;
	}

	PBGProviderList = make(map[string] PBGProvider, 0);
	PBGProviderBackDepList = make(map [string] []string, 0);
	PBGProviderForwardDepList = make(map [string] []string, 0);
	PBGProviderDistCache = make(map [PBGProviderDistCacheEntry] int, 0);
	PBGProviderListMutex = &sync.Mutex{}
}

// Locked version of getDepDistance. Does the actual calculation.
func lockedGetDepDistance(from string, to string, curr int) (int, bool) {
	pair := PBGProviderDistCacheEntry { from, to };

	if _, ok := PBGProviderDistCache[pair]; ok {
		return curr + PBGProviderDistCache[pair], true;
	}

	chans := make([](chan int), 0);

	for _, dep := range PBGProviderBackDepList[from] {
		if dep == to {
			return curr, false;
		}

		// Recursively and parallel-ly check its dependencies
		ch := make(chan int, 0);

		go func () {
			res, _ := lockedGetDepDistance(dep, to, curr + 1);
			log.Printf("Getting dep from %s to %s\n", dep, to);
			ch <- res
			close(ch);
		}();

		chans = append(chans, ch);
	}

	// Wait for the results and take the max.
	max := -1;

	for _, ch := range chans {
	 	for msg := range ch {
			if msg > max {
				max = msg;
			}
		}
	}

	return max, false;
}

// Calculate the amount of links between two providers via dependencies.
// Used internally for a topographic sort of the providers to create a safe
// execution order.
func getDepDistance(from string, to string) int {
	PBGProviderListMutex.Lock();

	res, cached := lockedGetDepDistance(from, to, 0);

	// Store newly computed results
	if !cached {
		PBGProviderDistCache[PBGProviderDistCacheEntry {from, to}] = res;
	}

	PBGProviderListMutex.Unlock();

	return res;
}

func (pbg *ProgramBehaviorGraph) SetProviderOptions(provider string, options map[string] interface{}) {
	pbg.options[provider] = options;
}

// Execute the currently installed providers 
func ExecuteProviders(pbg *ProgramBehaviorGraph, whitelist []string) {
	roots := getProviderRoots()

	if len(roots) == 0 {
		panic("No root providers!");
	}

	log.Printf("Found %s are root providers\n", roots)

	providerQueue := make(PriorityQueue, len(PBGProviderList));

	i := 0

	// Load the providers into a priority queue via their distance to roots.
	for dep, _ := range PBGProviderList {
		maxRoot := -1

		// Find the longest distance to any of the roots
		for _, root := range roots {
			dist := getDepDistance(dep, root) + 1;

			if dist > maxRoot {
				maxRoot = dist;
			} 
		}

		if maxRoot == -1 && len(PBGProviderBackDepList[dep]) > 0 {
			panic(fmt.Sprintf("no path to any roots from %s", dep));
		}

		log.Printf("Found %s has %d hops to root (%d dependencies)\n", dep, maxRoot, len(PBGProviderBackDepList[dep]));

		providerQueue[i] = &Item{
			value: dep,
			priority: maxRoot,
			index: i,
		};

		i++;
	}

	heap.Init(&providerQueue);

	// Execute the providers by lowest distance order. This means 
	// roots first and then their immediate dependents, etc.
	for providerQueue.Len() > 0 {
		dep := heap.Pop(&providerQueue).(*Item).value;
		opt, _ := pbg.options[dep];

		if len(whitelist) > 0 && sort.SearchStrings(whitelist, dep) >= len(whitelist) {
			log.Printf("Skipping stage %s for not being in whitelist...\n", dep);
			continue
		}

		log.Printf("Executing %s...\n", dep);
		start := time.Now()
		log.SetPrefix("[" + strings.ToUpper(dep) + "] ")
		PBGProviderList[dep](pbg, opt);
		log.SetPrefix("[PBG] ")
		end := time.Now()
		elapsed := end.Sub(start)

		log.Printf("Finished %s in %s", dep, elapsed.String())
	}
}


func RegisterProvider(name string, prov PBGProvider, deps ...string) {
	// Unsure if this is going to be called before we init so just try anyway
	initProviderObjects();

	PBGProviderListMutex.Lock();

	if PBGProviderList[name] != nil {
		panic("name already registered: " + name);
	}

	PBGProviderList[name] = prov;
	PBGProviderBackDepList[name] = deps;

	for _, dep := range deps {
		if PBGProviderForwardDepList[dep] != nil {
			PBGProviderForwardDepList[dep] = make([]string, 0);
		}

		PBGProviderForwardDepList[dep] = append(PBGProviderForwardDepList[dep], name);
	}

	PBGProviderListMutex.Unlock();
}


func init() {
	initProviderObjects();
}