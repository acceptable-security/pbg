#!/usr/bin/env python3

import os
import subprocess
import sys

if __name__ == "__main__":
	if not "DYNAMORIO_HOME" in os.environ:
		print("Failed to find DYNAMORIO_HOME in environment")
		sys.exit(1)

	if len(sys.argv) < 2:
		print('Usage: python3', sys.argv[0], "cmd [args...]")
		sys.exit(1)

	home = os.environ['DYNAMORIO_HOME']
	drrun = os.path.join(home, "bin64/drrun")

	print(home, drrun)

	instrace_folder = os.path.join(home, "../build_samples/bin/")
	instrace = os.path.join(instrace_folder, "libinstrace_x86_text.so")

	output = subprocess.check_output(
		drrun + ' -c ' + instrace +  ' -verbose 5 -- ' + ' '.join(sys.argv[1:]),
		stderr = subprocess.STDOUT,
		shell = True
	)

	lines = str(output).split('\\n')

	for line in lines:
		if not "Data file" in line:
			continue

		log_path = line[len("Data file "):]
		log_path = log_path[:-len(" created")]

		f = open(log_path).read()
		print(f)
		os.unlink(log_path)