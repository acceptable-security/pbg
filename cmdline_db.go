package main

import (
	"fmt"
	"os"
	"pbg/graph"
	"flag"
)

func dbUsage() {
	fmt.Printf("Usage: %s database [init -db file.db] [add -db file.db -s .. -v .. -o ..] [query -db file.db -cmd ...]\n", os.Args[0])
	os.Exit(1)
}

func dbInitCmd() {
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	initFile := initCmd.String("db", "", "name of the database")
	initCmd.Parse(os.Args[3:])

	_, err := graph.NewPBG("leveldb", *initFile, true)

	if err != nil {
		panic(err)
	}
}

func dbAddCmd() {
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	addFile := addCmd.String("db", "", "name of the database")
	addSubject := addCmd.String("s", "", "subject of the triplet")
	addVerb := addCmd.String("v", "", "verb of the triplet")
	addObject := addCmd.String("o", "", "object of the triplet")
	addCmd.Parse(os.Args[3:])

	pbg, err := graph.NewPBG("leveldb", *addFile, false)

	if err != nil {
		panic(err)
	}

	pbg.AddRelation(*addSubject, *addVerb, *addObject);
}

func dbQueryCmd() {
	queryCmd := flag.NewFlagSet("query", flag.ExitOnError)
	queryFile := queryCmd.String("db", "", "name of the database")
	queryString := queryCmd.String("cmd", "", "command to execute")
	queryCmd.Parse(os.Args[3:]);

	pbg, err := graph.NewPBG("leveldb", *queryFile, false)

	if err != nil {
		panic(err)
	}

	results, err := pbg.Query(*queryString)

	if err != nil {
		panic(err)
	}

	for i, res := range results {
		fmt.Printf("%d: %s\n", i, res)
	}
}

func databaseCmd() {
	if len(os.Args) < 3 {
		dbUsage()
	}

	switch os.Args[2] {
	case "init": dbInitCmd()
	case "add": dbAddCmd()
	case "query": dbQueryCmd()
	default: dbUsage()
	}
}
