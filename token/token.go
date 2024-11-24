package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	PROCEDURE = "PROCEDURE"
	IS        = "IS"
	BEGIN     = "BEGIN"
	END       = "END"
	PROGRAM   = "PROGRAM"

	ASSIGN    = ":="
	SEMICOLON = ";"

	IF    = "IF"
	THEN  = "THEN"
	ELSE  = "ELSE"
	ENDIF = "ENDIF"

	WHILE    = "WHILE"
	DO       = "DO"
	ENDWHILE = "ENDWHILE"

	REPEAT = "REPEAT"
	UNTIL  = "UNTIL"

	FOR    = "FOR"
	FROM   = "FROM"
	TO     = "TO"
	DOWNTO = "DOWNTO"
	ENDFOR = "ENDFOR"

	READ  = "READ"
	WRITE = "WRITE"

	LPAREN = "("
	RPAREN = ")"
	COMMA  = ","

	LBRACKET = "["
	RBRACKET = "]"

	COLON = ":"

	T = "T"

	PLUS   = "+"
	MINUS  = "-"
	MULT   = "*"
	DIVIDE = "/"
	MODULO = "%"

	EQUALS  = "="
	NEQUALS = "!="
	GE      = ">"
	LE      = "<"
	GEQ     = ">="
	LEQ     = "<="

	COMMENT = "#"

	NUM         = "NUM"
	PIDENTIFIER = "PIDENTIFIER"
)

var TokenMap = map[TokenType]int{
	ILLEGAL:     1,
	EOF:         2,
	PROCEDURE:   3,
	IS:          4,
	BEGIN:       5,
	END:         6,
	PROGRAM:     7,
	ASSIGN:      8,
	SEMICOLON:   9,
	IF:          10,
	THEN:        11,
	ELSE:        12,
	ENDIF:       13,
	WHILE:       14,
	DO:          15,
	ENDWHILE:    16,
	REPEAT:      17,
	UNTIL:       18,
	FOR:         19,
	FROM:        20,
	TO:          21,
	DOWNTO:      22,
	ENDFOR:      23,
	READ:        24,
	WRITE:       25,
	LPAREN:      26,
	RPAREN:      27,
	COMMA:       28,
	LBRACKET:    29,
	RBRACKET:    30,
	COLON:       31,
	T:           32,
	PLUS:        33,
	MINUS:       34,
	MULT:        35,
	DIVIDE:      36,
	MODULO:      37,
	EQUALS:      38,
	NEQUALS:     39,
	GE:          40,
	LE:          41,
	GEQ:         42,
	LEQ:         43,
	COMMENT:     44,
	NUM:         45,
	PIDENTIFIER: 46,
}

var keywords = map[string]TokenType{
	"PROCEDURE": PROCEDURE,
	"IS":        IS,
	"BEGIN":     BEGIN,
	"END":       END,
	"PROGRAM":   PROGRAM,
	"IF":        IF,
	"THEN":      THEN,
	"ELSE":      ELSE,
	"ENDIF":     ENDIF,
	"WHILE":     WHILE,
	"DO":        DO,
	"ENDWHILE":  ENDWHILE,
	"REPEAT":    REPEAT,
	"UNTIL":     UNTIL,
	"FOR":       FOR,
	"FROM":      FROM,
	"TO":        TO,
	"ENDFOR":    ENDFOR,
	"READ":      READ,
	"WRITE":     WRITE,
	"T":         T,
}

func LookupKeyword(ident string) (TokenType, bool) {
	tok, ok := keywords[ident]
	return tok, ok
}
