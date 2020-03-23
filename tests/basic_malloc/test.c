int main(void) {
	char* test = malloc(64);

	for ( int i = 0; i < 64; i++ ) {
		test[i] = 0x41;
	}

	free(test);
}