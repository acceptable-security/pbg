// g.V("main")
//  .Out("has-var")
//  .Out("decl-at")
//  .Out("line-content")
//  .All()

g.V().Tag("subject").Out(null, "predicate").Tag("object").All()
// g.V().OutPredicates().All()
// g.V().In("miss-address").Out("next-address").Limit(500).All()
/// g.V().In("text-at-pc").Limit(500).All()
// g.V().Out("next-address").All()

