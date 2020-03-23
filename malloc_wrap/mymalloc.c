// Source originally from
// http://www.cs.cmu.edu/afs/cs/academic/class/15213-s03/src/interposition/mymalloc.c
// Adapted to its current form by Avi Saven

#define _USE_GNU
#include <stdio.h>
#include <dlfcn.h>

static void* (*mallocp)(size_t size) = NULL;
static void (*freep)(void *) = NULL;

// TODO: print 

static inline
void init_address() {
	char *error;

	if ( !mallocp ) {
		mallocp = dlsym(RTLD_NEXT, "malloc");

		if ( (error = dlerror()) != NULL ) {
			fputs(error, stderr);
			exit(1);
		}
	}

	if ( !freep ) {
		freep = dlsym(RTLD_NEXT, "free");

		if ( (error = dlerror()) != NULL ) {
			fputs(error, stderr);
			exit(1);
		}
	}
}

void* malloc(size_t size) {
	init_address();
	void *ptr = mallocp(size);
	printf("m %zu %p\n", size, ptr);

	return ptr;
}

void free(void* ptr) {
	init_address();
	printf("f %p\n", ptr);     
	freep(ptr);
}