package parser

import "github.com/Meduza3/imp/lexer"

type Parser struct {
	l *lexer.Lexer
}

func New(l *lexer.Lexer) *Parser {
	return &Parser{l}
}
