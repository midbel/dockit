package repr

type InspectableExpr interface {
	Id() string
	Kind() string
	Params() map[string]string
	Children() []InspectableExpr
}

type Visitor interface {
	Visit(InspectableExpr)
}
