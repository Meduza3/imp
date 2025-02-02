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
	// defer func() {
	// 	fmt.Println(tok)
	// }()
	switch l.ch {
	case '+':
		tok = l.newToken(token.PLUS, l.ch)
	case '!':
		if l.peekChar() == '=' {
			tok = token.Token{Type: token.NEQUALS, Literal: "!=", Line: l.currentLine}
		}
		l.readChar()
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
			tok = l.newToken(token.GR, l.ch)
		}
	case ';':
		tok = l.newToken(token.SEMICOLON, l.ch)
	case '#':
		l.skipComment()
		return l.NextToken()
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
	input = `
	PROCEDURE built_in_mult() IS
  temp, left_sign
BEGIN

    IF built_in_left <= 0 THEN
      built_in_right:=0-built_in_right;
      built_in_left:=0-built_in_left;
    ENDIF

    built_in_result:=0;

    REPEAT
    temp:=built_in_left/2;
    temp:=temp+temp;
    IF temp != built_in_left THEN
    built_in_result:=built_in_result+built_in_right;
    ENDIF
    built_in_right:=built_in_right+built_in_right;
    built_in_left:=built_in_left/2;
    UNTIL built_in_left = 0;

END

PROCEDURE built_in_div() IS
    multiple, result_sign, temp
BEGIN
		IF built_in_right = 0 THEN
			built_in_result := 0;
		ELSE
    IF built_in_left <= 0 THEN
    built_in_left:= 0 - built_in_left;
    result_sign := 1;
    ELSE
    result_sign := 0;
    ENDIF

    built_in_result := 0;
    multiple := 1;

    IF built_in_right <= 0 THEN
    built_in_right:= 0 - built_in_right;
    result_sign:= 1 - result_sign;
    ENDIF

    REPEAT
        multiple:= multiple + multiple;
        built_in_right:= built_in_right + built_in_right;
    UNTIL built_in_right >= built_in_left;

    REPEAT
        IF built_in_left >= built_in_right THEN
            built_in_left := built_in_left - built_in_right;
            built_in_result := built_in_result + multiple;
        ENDIF
        built_in_right := built_in_right / 2;
        multiple := multiple / 2;
    UNTIL multiple = 0;

    IF result_sign != 0 THEN
        IF built_in_left != 0 THEN
            built_in_result:=-1-built_in_result;
        ELSE
            built_in_result:=0-built_in_result;
        ENDIF
    ENDIF
		ENDIF
END
PROCEDURE built_in_mod() IS
    current_divisor, dividend_sign, divisor_sign
BEGIN

    IF built_in_left <= 0 THEN
        built_in_left:=0-built_in_left;
        dividend_sign:=1;
    ELSE
        dividend_sign:=0;
    ENDIF

    IF built_in_right <= 0 THEN
        built_in_right:=0-built_in_right;
        divisor_sign:=1;
    ELSE
        divisor_sign:=0;
    ENDIF

    current_divisor:=built_in_right;

    REPEAT
        current_divisor:=current_divisor+current_divisor;
    UNTIL current_divisor > built_in_left;

    REPEAT
        current_divisor := current_divisor / 2;
        IF built_in_left >= current_divisor THEN
            built_in_left := built_in_left - current_divisor;
        ENDIF
    UNTIL built_in_left < built_in_right;

    built_in_result:=built_in_left;

    IF built_in_result != 0 THEN

    IF dividend_sign != 0 THEN
        built_in_result:=built_in_right-built_in_result;
    ENDIF

    IF divisor_sign != 0 THEN
        built_in_result:=built_in_result-built_in_right;
    ENDIF

    ENDIF
END ` + input
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
