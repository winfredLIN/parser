package parser

import "strings"

type Block interface {
	MatchBegin(tokenType int, token *yySymType) bool
	MatchEnd(tokenType int, token *yySymType) bool
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

func (b BeginEndBlock) MatchBegin(tokenType int, token *yySymType) bool {
	return tokenType == begin
}

func (b BeginEndBlock) MatchEnd(tokenType int, token *yySymType) bool {
	return true
}

type IfEndIfBlock struct{}

func (b IfEndIfBlock) MatchBegin(tokenType int, token *yySymType) bool {
	return tokenType == ifKwd
}

func (b IfEndIfBlock) MatchEnd(tokenType int, token *yySymType) bool {
	return tokenType == ifKwd
}

type CaseEndCaseBlock struct{}

func (b CaseEndCaseBlock) MatchBegin(tokenType int, token *yySymType) bool {
	return tokenType == caseKwd
}

func (b CaseEndCaseBlock) MatchEnd(tokenType int, token *yySymType) bool {
	return tokenType == caseKwd
}

type RepeatEndRepeatBlock struct{}

func (b RepeatEndRepeatBlock) MatchBegin(tokenType int, token *yySymType) bool {
	return tokenType == repeat
}

func (b RepeatEndRepeatBlock) MatchEnd(tokenType int, token *yySymType) bool {
	return tokenType == repeat
}

type WhileEndWhileBlock struct{}

func (b WhileEndWhileBlock) MatchBegin(tokenType int, token *yySymType) bool {
	return tokenType == identifier && strings.ToUpper(token.ident) == "WHILE"
}

func (b WhileEndWhileBlock) MatchEnd(tokenType int, token *yySymType) bool {
	return tokenType == identifier && strings.ToUpper(token.ident) == "WHILE"
}

type BeginEndBlock struct{}

func (b LoopEndLoopBlock) MatchBegin(tokenType int, token *yySymType) bool {
	return tokenType == identifier && strings.ToUpper(token.ident) == "LOOP"
}

func (b LoopEndLoopBlock) MatchEnd(tokenType int, token *yySymType) bool {
	return tokenType == identifier && strings.ToUpper(token.ident) == "LOOP"
}
