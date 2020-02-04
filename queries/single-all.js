// g.V("main")
//  .Out("has-var")
//  .Out("decl-at")
//  .Out("line-content")
//  .All()

g.V().Tag("subject").Out(null, "predicate").Tag("object").All()