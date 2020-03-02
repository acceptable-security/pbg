function resolveTypeName(typeId) {
	var type = g.V(typeId).Out("has-type-name").ToArray()[0]

	if ( type ) {
		return type
	}

	var type = g.V(typeId).OutPredicates().ToArray()[0]
	var next = g.V(typeId).Out(type).ToArray()[0]

	if ( type == "pointer-type" ) {
		return resolveTypeName(next) + "*"
	} else if (type == "const-type" ) {
		return "const " + resolveTypeName(next)
	} else if (type == "restrict-type" ) {
		return "restrict " + resolveTypeName(next)
	} else {
		throw new Error("Unknown type id " + typeId + " (" + type + ")");
	}
}

variables = g.V('main').Out("has-var").ToArray()

var folder = "tests/tcc/tcc-0.9.27/"
var funcFile = g.V("main").In().ToArray()[0];

for ( var i = 0; i < variables.length; i++ ) {
	var lineLoc = g.V(variables[i]).Out("decl-at").ToArray()[0]
	var line = g.V(funcFile + ':' + lineLoc).Out("line-content").ToArray()[0]
	var typeId = g.V(variables[i]).Out("has-var-type").ToArray()[0]
	var type = resolveTypeName(typeId)

	g.Emit(variables[i] + "(" + type + ") on " + lineLoc + ": " + line)
}
