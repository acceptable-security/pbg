package main

import (
	"log"
	"flag"
	"pbg/graph"
	"encoding/json"
	"os"
	"io/ioutil"
	"fmt"
	"sort"
	"strings"
)

func projUsage() {
	log.Printf("Usage: %s project [create -config=file.json -db=database -backend=backend] [query -db=database -backend=backend -query=file.js -draw=output.png]\n", os.Args[0]);
	os.Exit(1);
}

func projCreateCmd() {
	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	createFile := createCmd.String("config", "", "configuation file path")
	createDb := createCmd.String("db", "", "database file path")
	createBackend := createCmd.String("backend", "leveldb", "database backend")
	createWhitelist := createCmd.String("whitelist", "", "pass whitelist")

	createCmd.Parse(os.Args[3:])

	file, err := ioutil.ReadFile(*createFile)

	if err != nil {
		panic(err);
	}

	options := make(map[string] map[string] interface {})

	if err := json.Unmarshal(file, &options); err != nil {
		panic(err)
	}

	pbg, err := graph.NewPBG(*createBackend, *createDb, true)

	if err != nil {
		panic(err)
	}

	for key, val := range options {
		pbg.SetOptions(key, val)
	}

	whitelist := strings.Split(*createWhitelist, ",")
	sort.Strings(whitelist)

	graph.ExecuteProviders(pbg, whitelist)
}

func projQueryCmd() {
	queryCmd := flag.NewFlagSet("query", flag.ExitOnError)
	queryDb := queryCmd.String("db", "", "database file path")
	queryBackend := queryCmd.String("backend", "leveldb", "database backend")
	queryQuery := queryCmd.String("query", "", "query file path")
	queryDraw := queryCmd.String("draw", "", "picture to draw output of")

	queryCmd.Parse(os.Args[3:])

	pbg, err := graph.NewPBG(*queryBackend, *queryDb, false)

	if err != nil {
		panic(err)
	}

	queryString, err := ioutil.ReadFile(*queryQuery)

	if err != nil {
		panic(err)
	}

	if *queryDraw == "" {
		results, err := pbg.Query(string(queryString))

		if err != nil {
			panic(err)
		}

		for i, res := range results {
			log.Printf("%d: %s\n", i, res)
		}			
	} else {
		dot := pbg.Draw(string(queryString))

		fmt.Println(dot)
	}
}


func projectCmd() {
	if len(os.Args) < 3 {
		projUsage();
	}


	switch os.Args[2] {
	case "create":
		projCreateCmd();
	case "query":
		projQueryCmd()
	default:
		projUsage();
	}
}