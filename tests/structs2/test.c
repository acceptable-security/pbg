int main(void) {
	struct {
		struct { int foo; int bar; } blah;
		int x;
		int y;
	} z;

	z.x = 1;
	z.y = 2;

	return z.x * z.y;
}

