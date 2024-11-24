package lexer

import (
	"github.com/Meduza3/imp/parser"
	"github.com/Meduza3/imp/token"
)

type YyLex struct {
	input        string
	position     int
	readPosition int
	ch           byte
	currentToken *token.Token
}

func (l *YyLex) Lex(yylval *parser.YySymType) int {
	l.skipWhitespace()

	switch l.ch {
	case '+':
		l.currentToken = newToken(token.PLUS, l.ch)
		l.readChar()
	case '(':
		l.currentToken = newToken(token.LPAREN, l.ch)
		l.readChar()
	case ')':
		l.currentToken = newToken(token.RPAREN, l.ch)
		l.readChar()
	case ',':
		l.currentToken = newToken(token.COMMA, l.ch)
		l.readChar()
	case '[':
		l.currentToken = newToken(token.LBRACKET, l.ch)
		l.readChar()
	case ']':
		l.currentToken = newToken(token.RBRACKET, l.ch)
		l.readChar()
	case ':':
		l.readChar()
		if l.ch == '=' {
			l.currentToken = &token.Token{Type: token.ASSIGN, Literal: ":="}
			l.readChar()
		} else {
			l.currentToken = newToken(token.COLON, ':')
		}
	case '-':
		l.currentToken = newToken(token.MINUS, l.ch)
		l.readChar()
	case '*':
		l.currentToken = newToken(token.MULT, l.ch)
		l.readChar()
	case '/':
		l.currentToken = newToken(token.DIVIDE, l.ch)
		l.readChar()
	case '%':
		l.currentToken = newToken(token.MODULO, l.ch)
		l.readChar()
	case '=':
		l.currentToken = newToken(token.EQUALS, l.ch)
		l.readChar()
	case '<':
		l.readChar()
		if l.ch == '=' {
			l.currentToken = &token.Token{Type: token.LEQ, Literal: "<="}
			l.readChar()
		} else {
			l.currentToken = newToken(token.LE, '<')
		}
	case '>':
		l.readChar()
		if l.ch == '=' {
			l.currentToken = &token.Token{Type: token.GEQ, Literal: ">="}
			l.readChar()
		} else {
			l.currentToken = newToken(token.GE, '>')
		}
	case ';':
		l.currentToken = newToken(token.SEMICOLON, l.ch)
		l.readChar()
	case '#':
		l.currentToken = newToken(token.COMMENT, l.ch)
		l.skipComment()
	case 0:
		literal := "<$EOF$>"
		l.currentToken = &token.Token{Type: token.EOF, Literal: literal}
	default:
		if isDigit(l.ch) {
			literal := l.readNumber()
			l.currentToken = &token.Token{Type: token.NUM, Literal: literal}
		} else if isLowercaseLetter(l.ch) {
			literal := l.readPidentifier()
			l.currentToken = &token.Token{Type: token.PIDENTIFIER, Literal: literal}
		} else if isUppercaseLetter(l.ch) {
			literal := l.readKeyword()
			kwToken, ok := token.LookupKeyword(literal)
			if !ok {
				l.currentToken = newToken(token.ILLEGAL, l.ch)
				yylval.Token = l.currentToken
				return token.TokenMap[l.currentToken.Type]
			}
			l.currentToken = &token.Token{Type: kwToken, Literal: literal}
		} else {
			l.currentToken = newToken(token.ILLEGAL, l.ch)
			l.readChar()
		}
	}

	yylval.Token = l.currentToken
	return token.TokenMap[l.currentToken.Type]
}

func newToken(tokenType token.TokenType, ch byte) *token.Token {
	return &token.Token{Type: tokenType, Literal: string(ch)}
}

func New(input string) *YyLex {
	l := &YyLex{input: input}
	l.readChar()
	return l
}

func (l *YyLex) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *YyLex) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *YyLex) readPidentifier() string {
	position := l.position
	for isLowercaseLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *YyLex) readKeyword() string {
	position := l.position
	for isUppercaseLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *YyLex) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *YyLex) skipComment() {
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

func (l *YyLex) Error(s string) {
	// Error handling implementation
}
