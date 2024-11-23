package lexer

type Lexer struct {
	currentChar int
}

func (l *Lexer) NextChar() {
	l.currentChar++
}
