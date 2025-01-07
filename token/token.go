package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
	Line    int
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
