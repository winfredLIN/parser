package parser

import "strings"

type Block interface {
	MatchBegin(token *Token) bool
	MatchEnd(token *Token) bool
}

type Blocker struct {
	Scanner *Scanner
}

func NewBlocker() *Blocker {
	return &Blocker{}
}

var allBlocks []Block = []Block{
	BeginEndBlock{},
	IfEndIfBlock{},
	CaseEndCaseBlock{},
	RepeatEndRepeatBlock{},
	WhileEndWhileBlock{},
	LoopEndLoopBlock{},
}

type LoopEndLoopBlock struct{}

func (b BeginEndBlock) MatchBegin(token *Token) bool {
	return token.tokenType == begin
}

func (b BeginEndBlock) MatchEnd(token *Token) bool {
	return true
}

type IfEndIfBlock struct{}

func (b IfEndIfBlock) MatchBegin(token *Token) bool {
	return token.tokenType == ifKwd
}

func (b IfEndIfBlock) MatchEnd(token *Token) bool {
	return token.tokenType == ifKwd
}

type CaseEndCaseBlock struct{}

func (b CaseEndCaseBlock) MatchBegin(token *Token) bool {
	return token.tokenType == caseKwd
}

func (b CaseEndCaseBlock) MatchEnd(token *Token) bool {
	return token.tokenType == caseKwd
}

type RepeatEndRepeatBlock struct{}

func (b RepeatEndRepeatBlock) MatchBegin(token *Token) bool {
	return token.tokenType == repeat
}

func (b RepeatEndRepeatBlock) MatchEnd(token *Token) bool {
	return token.tokenType == repeat
}

type WhileEndWhileBlock struct{}

func (b WhileEndWhileBlock) MatchBegin(token *Token) bool {
	return token.tokenType == identifier && strings.ToUpper(token.tokenValue.ident) == "WHILE"
}

func (b WhileEndWhileBlock) MatchEnd(token *Token) bool {
	return token.tokenType == identifier && strings.ToUpper(token.tokenValue.ident) == "WHILE"
}

type BeginEndBlock struct{}

func (b LoopEndLoopBlock) MatchBegin(token *Token) bool {
	return token.tokenType == identifier && strings.ToUpper(token.tokenValue.ident) == "LOOP"
}

func (b LoopEndLoopBlock) MatchEnd(token *Token) bool {
	return token.tokenType == identifier && strings.ToUpper(token.tokenValue.ident) == "LOOP"
}
