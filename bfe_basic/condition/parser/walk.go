package parser

type Visitor interface {
	Visit(Node) Visitor
}

func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	switch n := node.(type) {
	case *BinaryExpr:
		Walk(v, n.X)
		Walk(v, n.Y)
	case *UnaryExpr:
		Walk(v, n.X)
	case *Ident:
	case *CallExpr:
		Walk(v, n.Fun)
		Walk(v, n.Args)
	case *ParenExpr:
		Walk(v, n.X)
	}
}

type inspector func(Node) bool

func (f inspector) Visit(node Node) Visitor {
	if f(node) {
		return f
	}

	return nil
}

func Inspect(node Node, f func(Node) bool) {
	Walk(inspector(f), node)
}
