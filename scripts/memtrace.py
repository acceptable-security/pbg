#!/usr/bin/env python3

import gzip
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

	try:
		output = subprocess.check_output(
			drrun + ' -t drcachesim -verbose 6 -LL_miss_file ./.tmp_cache.gz -- ' + ' '.join(sys.argv[1:]),
			stderr = subprocess.STDOUT,
			shell = True
		)

		print("pc,cmd,addr")

		for line in str(output).split('\\n'):
			line = line.strip()

			if line[:2] != '::' or not '@' in line:
				continue

			cmd = line[line.index('@'):].split(' ')

			if len(cmd) < 2 or not cmd[1] in [ 'read', 'write' ]:
				continue

			address = cmd[0]
			destination = cmd[2]

			print(','.join([address, cmd[1], destination]))

		f = gzip.open('./.tmp_cache.gz', 'rb')
		cache_data = f.read()
		f.close()
		os.unlink("./.tmp_cache.gz")

		f = open('cache_miss.csv', 'w')
		f.write("pc,addr\n")
		f.write(cache_data.decode('ascii'))
		f.close()
	except subprocess.CalledProcessError as e:
		print("Failed:", e.output, file=sys.stderr)
