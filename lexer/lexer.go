package lexer

import (
	"github.com/Meduza3/imp/token"
)

type Lexer struct {
	input        string
	position     int //current position
	readPosition int // position + 1
	ch           byte
	currentLine  int
}

func (l *Lexer) NextToken() token.Token {
	l.skipWhitespace()
	var tok token.Token
	switch l.ch {
	case '+':
		tok = l.newToken(token.PLUS, l.ch)
	case '(':
		tok = l.newToken(token.LPAREN, l.ch)
	case ')':
		tok = l.newToken(token.RPAREN, l.ch)
	case ',':
		tok = l.newToken(token.COMMA, l.ch)
	case '[':
		tok = l.newToken(token.LBRACKET, l.ch)
	case ']':
		tok = l.newToken(token.RBRACKET, l.ch)
	case ':':
		peeked := l.peekChar()
		if peeked == '=' {
			l.readChar()
			tok = token.Token{Type: token.ASSIGN, Literal: ":=", Line: l.currentLine}
		} else {
			tok = l.newToken(token.COLON, ':')
		}
	case '-':
		tok = l.newToken(token.MINUS, l.ch)
	case '*':
		tok = l.newToken(token.MULT, l.ch)
	case '/':
		tok = l.newToken(token.DIVIDE, l.ch)
	case '%':
		tok = l.newToken(token.MODULO, l.ch)
	case '=':
		tok = token.Token{Type: token.EQUALS, Literal: "="}
	case '<':
		peeked := l.peekChar()
		if peeked == '=' {
			l.readChar()
			tok = token.Token{Type: token.LEQ, Literal: "<="}
		} else {
			tok = l.newToken(token.LE, l.ch)
		}
	case '>':
		peeked := l.peekChar()
		if peeked == '=' {
			l.readChar()
			tok = token.Token{Type: token.GEQ, Literal: ">=", Line: l.currentLine}
		} else {
			tok = l.newToken(token.GE, l.ch)
		}
	case ';':
		tok = l.newToken(token.SEMICOLON, l.ch)
	case '#':
		l.skipComment()
	case 0:
		literal := "<$EOF$>"
		tok.Type = token.EOF
		tok.Literal = literal
	default:
		if isDigit(l.ch) {
			literal := l.readNumber()
			tok = token.Token{Type: token.NUM, Literal: literal, Line: l.currentLine}
			return tok
		} else if isLowercaseLetter(l.ch) {
			literal := l.readPidentifier()
			tok = token.Token{Type: token.PIDENTIFIER, Literal: literal, Line: l.currentLine}
			return tok
		} else if isUppercaseLetter(l.ch) {
			literal := l.readKeyword()
			tokenType, ok := token.LookupKeyword(literal)
			if !ok {
				tok = token.Token{Type: token.ILLEGAL, Literal: literal, Line: l.currentLine}
				return tok
			}
			tok = token.Token{Type: tokenType, Literal: literal, Line: l.currentLine}
			return tok
		} else {
			tok = l.newToken(token.ILLEGAL, l.ch)
		}
	}
	l.readChar()
	return tok
}

func (l *Lexer) newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch), Line: l.currentLine}
}

func New(input string) *Lexer {
	l := &Lexer{input: input, currentLine: 1}
	l.readChar()
	return l
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
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
		if l.ch == '\n' {
			l.currentLine++
		}
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
