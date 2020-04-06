#!/bin/bash

set -e

PBG=./pbg
TESTS=$(ls tests/)
TEST_ITERS=1
OUTPUTDIR="/localdisk/asaven-pbg/test_$(date +%Y-%m-%d_%H-%M-%S)"
BACKEND=sqlite
SOUFFLE=/localdisk/asaven-pbg/souffle-1.7.1/usr/bin/souffle

RESULTS="$OUTPUTDIR/results.txt"

echo "Validate parameters before continuing:"
echo ""
echo "The following tests will be run: $TESTS"
echo "The tests will be run for $TEST_ITERS iterations."
echo "The results are going to be written to $OUTPUTDIR"
echo "The tests will be written to a $BACKEND database"
echo "The queries will be executed with $SOUFFLE"
echo "The results will be written to $RESULTS"

read -t 60

echo "Starting in 5 seconds..."; sleep 5

mkdir -p $OUTPUTDIR
echo "NAME\tITER\tCREATE_TIME\tQUERY_TIME\tTEST_TIME" > $RESULTS


for i in $(seq 1 $TEST_ITERS); do
	echo "Test iteration $i"

	for test in tests/*; do
		# Name is the folder
		name=$(basename "$test")

		# Compute arguments
		db="$OUTPUTDIR/$name.db"
		rm -f $db
		conf="$test/$name.json"

		if [ ! -f "$conf" ]; then
			echo "$conf does not exist, skipping $name"
			continue
		fi

		echo "Creating project $name"
		TIMEFORMAT='%lU'

		create_time=$(time ( $PBG project create -db=$db -backend=$BACKEND -config=$conf ) 2>&1 1>$OUTPUTDIR/$name.create.$i.txt)

		# TODO: on validation switch i => f
		echo "Removing old datalog if present"
		rm -rf $OUTPUTDIR/datalog_dir/
		mkdir -p $OUTPUTDIR/datalog_dir/

		echo "Creating datalog directories"
		query_time=$(time ( $PBG project query -db=$db -backend=$BACKEND -datalog=$OUTPUTDIR/datalog_dir/ ) 2>&1 1>$OUTPUTDIR/$name.query.$i.txt)

		echo "Running memory tests"
		test_time=$(time ( $SOUFFLE --fact-dir=$OUTPUTDIR/datalog_dir -c -j64 ./queries/mem-test.dl) 2>&1 1>$OUTPUTDIR/$name.test.$i.txt)

		echo "Saving results"
		echo "$name\t$i\t$create_time\t$query_time\t$test_time" >> $RESULTS
	done
done