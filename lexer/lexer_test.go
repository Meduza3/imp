package lexer

import (
	"testing"

	"github.com/Meduza3/imp/parser"
	"github.com/Meduza3/imp/token"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		input     string
		tokenList []*token.Token
	}{
		{
			input: `# Binarna postać liczby
PROGRAM IS
n , p
BEGIN
READ n ;
REPEAT
p : = n / 2 ;
p : = 2 * p ;
IF n > p THEN
WRITE 1 ;
ELSE
	WRITE 0 ;
ENDIF
n : = n / 2 ;
UNTIL n = 0 ;
END`,
			tokenList: []*token.Token{
				{Type: token.COMMENT, Literal: "#"},
				{Type: token.PROGRAM, Literal: "PROGRAM"},
				{Type: token.IS, Literal: "IS"},
				{Type: token.PIDENTIFIER, Literal: "n"},
				{Type: token.COMMA, Literal: ","},
				{Type: token.PIDENTIFIER, Literal: "p"},
				{Type: token.BEGIN, Literal: "BEGIN"},
				{Type: token.READ, Literal: "READ"},
				{Type: token.PIDENTIFIER, Literal: "n"},
				{Type: token.SEMICOLON, Literal: ";"},
				{Type: token.REPEAT, Literal: "REPEAT"},
				{Type: token.PIDENTIFIER, Literal: "p"},
				{Type: token.COLON, Literal: ":"},
				{Type: token.EQUALS, Literal: "="},
				{Type: token.PIDENTIFIER, Literal: "n"},
				{Type: token.DIVIDE, Literal: "/"},
				{Type: token.NUM, Literal: "2"},
				{Type: token.SEMICOLON, Literal: ";"},
				{Type: token.PIDENTIFIER, Literal: "p"},
				{Type: token.COLON, Literal: ":"},
				{Type: token.EQUALS, Literal: "="},
				{Type: token.NUM, Literal: "2"},
				{Type: token.MULT, Literal: "*"},
				{Type: token.PIDENTIFIER, Literal: "p"},
				{Type: token.SEMICOLON, Literal: ";"},
				{Type: token.IF, Literal: "IF"},
				{Type: token.PIDENTIFIER, Literal: "n"},
				{Type: token.GE, Literal: ">"},
				{Type: token.PIDENTIFIER, Literal: "p"},
				{Type: token.THEN, Literal: "THEN"},
				{Type: token.WRITE, Literal: "WRITE"},
				{Type: token.NUM, Literal: "1"},
				{Type: token.SEMICOLON, Literal: ";"},
				{Type: token.ELSE, Literal: "ELSE"},
				{Type: token.WRITE, Literal: "WRITE"},
				{Type: token.NUM, Literal: "0"},
				{Type: token.SEMICOLON, Literal: ";"},
				{Type: token.ENDIF, Literal: "ENDIF"},
				{Type: token.PIDENTIFIER, Literal: "n"},
				{Type: token.COLON, Literal: ":"},
				{Type: token.EQUALS, Literal: "="},
				{Type: token.PIDENTIFIER, Literal: "n"},
				{Type: token.DIVIDE, Literal: "/"},
				{Type: token.NUM, Literal: "2"},
				{Type: token.SEMICOLON, Literal: ";"},
				{Type: token.UNTIL, Literal: "UNTIL"},
				{Type: token.PIDENTIFIER, Literal: "n"},
				{Type: token.EQUALS, Literal: "="},
				{Type: token.NUM, Literal: "0"},
				{Type: token.SEMICOLON, Literal: ";"},
				{Type: token.END, Literal: "END"},
				{Type: token.EOF, Literal: "<$EOF$>"},
			},
		},
		{
			// Add a test case for error handling
			input: "@invalid",
			tokenList: []*token.Token{
				{Type: token.ILLEGAL, Literal: "@"},
				{Type: token.PIDENTIFIER, Literal: "invalid"},
				{Type: token.EOF, Literal: "<$EOF$>"},
			},
		},
		{
			// Test case for all single-character operators
			input: "+ - * / % = < > : ;",
			tokenList: []*token.Token{
				{Type: token.PLUS, Literal: "+"},
				{Type: token.MINUS, Literal: "-"},
				{Type: token.MULT, Literal: "*"},
				{Type: token.DIVIDE, Literal: "/"},
				{Type: token.MODULO, Literal: "%"},
				{Type: token.EQUALS, Literal: "="},
				{Type: token.LE, Literal: "<"},
				{Type: token.GE, Literal: ">"},
				{Type: token.COLON, Literal: ":"},
				{Type: token.SEMICOLON, Literal: ";"},
				{Type: token.EOF, Literal: "<$EOF$>"},
			},
		},
	}

	for i, tt := range tests {
		l := New(tt.input)
		var tokens []*token.Token

		yylval := &parser.YySymType{}

		// Collect all tokens
		for {
			tokenType := l.Lex(yylval)
			tokens = append(tokens, yylval.Token)

			if tokenType == token.TokenMap[token.EOF] {
				break
			}
		}

		// Report token count mismatch but continue with comparison
		if len(tokens) != len(tt.tokenList) {
			t.Errorf("test[%d] - wrong number of tokens. expected=%d, got=%d",
				i, len(tt.tokenList), len(tokens))
		}

		// Compare tokens up to the length of the shorter list
		minLen := len(tokens)
		if len(tt.tokenList) < minLen {
			minLen = len(tt.tokenList)
		}

		for j := 0; j < minLen; j++ {
			if tokens[j].Type != tt.tokenList[j].Type {
				t.Errorf("test[%d][%d] - wrong token type. expected=%q, got=%q",
					i, j, tt.tokenList[j].Type, tokens[j].Type)
			}
			if tokens[j].Literal != tt.tokenList[j].Literal {
				t.Errorf("test[%d][%d] - wrong literal. expected=%q, got=%q",
					i, j, tt.tokenList[j].Literal, tokens[j].Literal)
			}
		}

		// Show any extra tokens in actual output
		if len(tokens) > len(tt.tokenList) {
			t.Errorf("test[%d] - extra tokens received:", i)
			for j := len(tt.tokenList); j < len(tokens); j++ {
				t.Errorf("  extra token[%d]: {Type: %q, Literal: %q}",
					j, tokens[j].Type, tokens[j].Literal)
			}
		}

		// Show any missing tokens that were expected
		if len(tt.tokenList) > len(tokens) {
			t.Errorf("test[%d] - missing expected tokens:", i)
			for j := len(tokens); j < len(tt.tokenList); j++ {
				t.Errorf("  missing token[%d]: {Type: %q, Literal: %q}",
					j, tt.tokenList[j].Type, tt.tokenList[j].Literal)
			}
		}
	}
}
