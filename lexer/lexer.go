package lexer

import (
	"github.com/Meduza3/imp/token"
)

type Lexer struct {
	input        string
	position     int //current position
	readPosition int // position + 1
	ch           byte
}

func (l *Lexer) NextToken() token.Token {
	l.skipWhitespace()
	var tok token.Token
	switch l.ch {
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
	// case ':':
	// 	l.readChar()
	// 	if l.ch == '=' {
	// 		l.currentToken = &token.Token{Type: token.ASSIGN, Literal: ":="}
	// 		l.readChar()
	// 	} else {
	// 		l.currentToken = newToken(token.COLON, ':')
	// 	}
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '*':
		tok = newToken(token.MULT, l.ch)
	case '/':
		tok = newToken(token.DIVIDE, l.ch)
	case '%':
		tok = newToken(token.MODULO, l.ch)
	case '=':
		tok = newToken(token.EQUALS, l.ch)
	// case '<':
	// 	l.readChar()
	// 	if l.ch == '=' {
	// 		l.currentToken = &token.Token{Type: token.LEQ, Literal: "<="}
	// 		l.readChar()
	// 	} else {
	// 		l.currentToken = newToken(token.LE, '<')
	// 	}
	// case '>':
	// 	l.readChar()
	// 	if l.ch == '=' {
	// 		l.currentToken = &token.Token{Type: token.GEQ, Literal: ">="}
	// 		l.readChar()
	// 	} else {
	// 		l.currentToken = newToken(token.GE, '>')
	// 	}
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case '#':
		l.skipComment()
	case 0:
		literal := "<$EOF$>"
		tok.Type = token.EOF
		tok.Literal = literal
	default:
		if isDigit(l.ch) {
			literal := l.readNumber()
			tok = token.Token{Type: token.NUM, Literal: literal}
		} else if isLowercaseLetter(l.ch) {
			literal := l.readPidentifier()
			tok = token.Token{Type: token.PIDENTIFIER, Literal: literal}
		}
	}
	l.readChar()
	return tok
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readPidentifier() string {
	position := l.position
	for isLowercaseLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readKeyword() string {
	position := l.position
	for isUppercaseLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	l.readChar() // Skip the newline character
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isLowercaseLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || ch == '_'
}

func isUppercaseLetter(ch byte) bool {
	return 'A' <= ch && ch <= 'Z'
}

func (l *Lexer) Error(s string) {
	// Error handling implementation
}
