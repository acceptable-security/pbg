package clang

import (
	"fmt"
	"modernc.org/cc/v2"
	"pbg/graph"
	"reflect"
)

// Create cc sources from file paths
func createSources(pbg *graph.ProgramBehaviorGraph) []cc.Source {
	sources := make([]cc.Source, 0);

	// All directories
	dirs, err := pbg.Query("g.V().In('contains-file').All()");

	if err != nil {
		panic(err);
	}

	fmt.Printf("Directories: %s\n", dirs);

	for _, dir := range dirs {
		dir = dir[1:len(dir) - 1]

		// All files from this directory
		query := fmt.Sprintf("g.V('%s').Out('contains-file').All()", dir)

		files, err := pbg.Query(query);

		if err != nil {
			panic(err);
		}

		fmt.Printf("Found files %s from %s (%s)\n", files, dir, reflect.TypeOf(dir));

		for _, file := range files {
			file = file[1:len(file) - 1];

			source, err := cc.NewFileSource(dir + "/" + file);

			if err != nil {
				panic(err);
			}

			sources = append(sources, source);
		}
	}

	return sources;
}

// Create an AST from the file sources
func createAST(sources []cc.Source, _includePaths []string, _sysIncludePaths []string) *cc.TranslationUnit {
	predef, includePaths, sysIncludePaths, err := cc.HostConfig("-std=c99")

	sources = append([]cc.Source { cc.NewStringSource("<predef>", predef) }, sources...)

	tu, err := cc.Translate(&cc.Tweaks{
		EnableAnonymousStructFields: true,
		EnableBinaryLiterals: true,
		EnableEmptyStructs: true,
		EnableImplicitBuiltins: true,
		EnableImplicitDeclarations: true,
		EnableOmitFuncDeclSpec: true,
		EnablePointerCompatibility: true,
		EnableReturnExprInVoidFunc: true,
		EnableTrigraphs: true,
		EnableUnionCasts: true,
		IgnoreUnknownPragmas: true,
		InjectFinalNL: true,
		TrackExpand:                 func(s string) { fmt.Println("expand " + s) },
		TrackIncludes:               func(s string) { fmt.Println("include " + s) },
	}, includePaths, sysIncludePaths, sources...);

	if err != nil {
		panic(fmt.Sprintf("Error: %v", err));
	}

	return tu;	
}

func walkFuncAST(pbg *graph.ProgramBehaviorGraph, funcDecl *cc.FunctionDefinition) {
	decl := funcDecl.Declarator
	declType := decl.Type
	fmt.Printf("Found type %v\n", declType)
}

func walkAST(pbg *graph.ProgramBehaviorGraph, tu *cc.TranslationUnit) {
	list := tu.ExternalDeclarationList

	for list != nil {
		extDecl := list.ExternalDeclaration

		if decl := extDecl.Declaration; decl != nil {

		} else if funcDecl := extDecl.FunctionDefinition; funcDecl != nil {

		} else {
			panic(fmt.Sprintf("Unknown decl %v", extDecl))
		}

		list = list.ExternalDeclarationList
	}
}

func loadElements(pbg *graph.ProgramBehaviorGraph, opt map [string] interface{}) {
	if _, ok := opt["doWork"]; !ok {
		return
	}

	sources := createSources(pbg);

	if len(sources) == 0 {
		fmt.Printf("No sources found, skipping ast...\n");
		return
	}

	_sysIncludePaths, ok := opt["sysIncludePaths"].([]interface {})
	sysIncludePaths := make([]string, 0)

	for _, path := range _sysIncludePaths {
		sysIncludePaths = append(sysIncludePaths, path.(string))
	}

	_includePaths, ok := opt["includePaths"].([]interface {})
	includePaths := make([]string, 0)

	if ok {
		for _, path := range _includePaths {
			includePaths = append(includePaths, path.(string))
		}
	}


	tu := createAST(sources, includePaths, sysIncludePaths);

	if tu == nil {
		fmt.Printf("failed to make ast\n")
		return
	}

	walkAST(pbg, tu)
}

func init() {
	graph.RegisterProvider("clang", loadElements, "files");
}