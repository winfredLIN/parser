package ast

import (
	"github.com/pingcap/parser/format"
)

type UnknownNode struct {
	node
}

func (n *UnknownNode) statement() {}

// Restore returns the sql text from ast tree
func (n *UnknownNode) Restore(ctx *format.RestoreCtx) error {
	ctx.WriteString(n.Text())
	return nil
}

func (n *UnknownNode) Accept(v Visitor) (node Node, ok bool) {
	return nil, false
}
