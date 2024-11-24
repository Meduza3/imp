%{
package parser

import (
	"github.com/Meduza3/imp/ast"
	"github.com/Meduza3/imp/token"
	"strconv"
)

func atoi(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		// Handle error appropriately
		return 0
	}
	return i
}

%}

%union {
	token       *token.Token
	program     *ast.ProgramAll
	procedures  []*ast.Procedure
	main        *ast.Main
	statements  []ast.Statement
	statement   ast.Statement
	expression  ast.Expression
	expressions []ast.Expression
	identifier  ast.Expression
	integer     int64
	condition   ast.Expression
	value       ast.Expression
	declarations []ast.Declaration
	arg_decls   []*ast.ArgDeclaration
	args        []ast.Expression
	proc_call   *ast.ProcedureCallStatement
	proc_head   *ast.ProcHead
	arg_decl    *ast.ArgDeclaration
}

%token <token> ILLEGAL
%token <token> EOF

%token <token> PROCEDURE IS BEGIN END PROGRAM
%token <token> ASSIGN SEMICOLON
%token <token> IF THEN ELSE ENDIF
%token <token> WHILE DO ENDWHILE
%token <token> REPEAT UNTIL
%token <token> FOR FROM TO DOWNTO ENDFOR
%token <token> READ WRITE
%token <token> LPAREN RPAREN COMMA
%token <token> LBRACKET RBRACKET COLON
%token <token> T
%token <token> PLUS MINUS MULT DIVIDE MODULO
%token <token> EQUALS NEQUALS GE LE GEQ LEQ
%token <token> NUM PIDENTIFIER

%start program_all

%type <program> program_all
%type <procedures> procedures
%type <main> main
%type <statements> commands
%type <statement> command
%type <expression> expression
%type <condition> condition
%type <value> value
%type <identifier> identifier
%type <declarations> declarations
%type <arg_decls> args_decl
%type <args> args
%type <proc_call> proc_call
%type <proc_head> proc_head


%%

program_all:
    procedures main
    {
        $$ = &ast.ProgramAll{
            Procedures: $1,
            Main:       $2,
        }
    }
  ;

procedures:
    procedures PROCEDURE proc_head IS declarations BEGIN commands END
    {
        $$ = append($1, &ast.Procedure{
            Token:        $2,
            Head:         $3,
            Declarations: $5,
            Commands:     $7,
        })
    }
  | procedures PROCEDURE proc_head IS BEGIN commands END
  {
        $$ = append($1, &ast.Procedure{
            Token:        $2,
            Head:         $3,
            Declarations: nil,
            Commands:     $6,
        })
    }
  |
  {
        $$ = []*ast.Procedure{}
    }
  ;

main:
    PROGRAM IS declarations BEGIN commands END
    {
        $$ = &ast.Main{
            Token:        $1,
            Declarations: $3,
            Commands:     $5,
        }
    }
  | PROGRAM IS BEGIN commands END
  {
        $$ = &ast.Main{
            Token:        $1,
            Declarations: nil,
            Commands:     $4,
        }
    }
  ;

commands:
    commands command
    {
        $$ = append($1, $2)
    }
  | command
  {
        $$ = []ast.Statement{$1}
    }
  ;

command:
    identifier ASSIGN expression SEMICOLON
    {
        $$ = &ast.AssignmentStatement{
            Token: $2,
            Left:  $1,
            Right: $3,
        }
    }
  | IF condition THEN commands ELSE commands ENDIF
  {
        $$ = &ast.IfStatement{
            Token:       $1,
            Condition:   $2,
            Consequence: $4,
            Alternative: $6,
        }
    }
  | IF condition THEN commands ENDIF
  {
        $$ = &ast.IfStatement{
            Token:       $1,
            Condition:   $2,
            Consequence: $4,
            Alternative: nil,
        }
    }
  | WHILE condition DO commands ENDWHILE
  {
        $$ = &ast.WhileStatement{
            Token:     $1,
            Condition: $2,
            Body:      $4,
        }
    }
  | REPEAT commands UNTIL condition SEMICOLON
  {
        $$ = &ast.RepeatStatement{
            Token:     $1,
            Body:      $2,
            Condition: $4,
        }
    }
  | FOR PIDENTIFIER FROM value TO value DO commands ENDFOR
  {
        $$ = &ast.ForStatement{
            Token:    $1,
            Iterator: &ast.Identifier{Token: $2, Value: $2.Literal},
            From:     $4,
            To:       $6,
            DownTo:   false,
            Body:     $8,
        }
    }
  | FOR PIDENTIFIER FROM value DOWNTO value DO commands ENDFOR
  {
        $$ = &ast.ForStatement{
            Token:    $1,
            Iterator: &ast.Identifier{Token: $2, Value: $2.Literal},
            From:     $4,
            To:       $6,
            DownTo:   true,
            Body:     $8,
        }
    }
  | proc_call SEMICOLON
  {
        $$ = $1
    }
  | READ identifier SEMICOLON
  {
        $$ = &ast.ReadStatement{
            Token:      $1,
            Identifier: $2,
        }
    }
  | WRITE value SEMICOLON
  {
        $$ = &ast.WriteStatement{
            Token: $1,
            Value: $2,
        }
    }
  ;

proc_head:
    PIDENTIFIER LPAREN args_decl RPAREN
    {
        $$ = &ast.ProcHead{
            Name:     &ast.Identifier{Token: $1, Value: $1.Literal},
            ArgsDecl: $3,
        }
    }
  ;

proc_call:
    PIDENTIFIER LPAREN args RPAREN
    {
        $$ = &ast.ProcedureCallStatement{
            Token:     $1,
            Name:      &ast.Identifier{Token: $1, Value: $1.Literal},
            Arguments: $3,
        }
    }
  ;

declarations:
    declarations COMMA PIDENTIFIER
    {
        ident := &ast.Identifier{Token: $3, Value: $3.Literal}
        $$ = append($1, &ast.VarDeclaration{
            Token: $3,
            Name:  ident,
        })
    }
  | declarations COMMA PIDENTIFIER LBRACKET NUM COLON NUM RBRACKET
  {
        fromVal := &ast.IntegerLiteral{Token: $5, Value: atoi($5.Literal)}
        toVal := &ast.IntegerLiteral{Token: $7, Value: atoi($7.Literal)}
        $$ = append($1, &ast.ArrayDeclaration{
            Token: $3,
            Name:  &ast.Identifier{Token: $3, Value: $3.Literal},
            From:  fromVal,
            To:    toVal,
        })
    }
  | PIDENTIFIER
  {
        $$ = []ast.Declaration{
            &ast.VarDeclaration{
                Token: $1,
                Name:  &ast.Identifier{Token: $1, Value: $1.Literal},
            },
        }
    }
  | PIDENTIFIER LBRACKET NUM COLON NUM RBRACKET
  {
        fromVal := &ast.IntegerLiteral{Token: $3, Value: atoi($3.Literal)}
        toVal := &ast.IntegerLiteral{Token: $5, Value: atoi($5.Literal)}
        $$ = []ast.Declaration{
            &ast.ArrayDeclaration{
                Token: $1,
                Name:  &ast.Identifier{Token: $1, Value: $1.Literal},
                From:  fromVal,
                To:    toVal,
            },
        }
    }
  ;

args_decl:
    args_decl COMMA PIDENTIFIER
    {
        arg := &ast.ArgDeclaration{
            Token: $3,
            Name:  &ast.Identifier{Token: $3, Value: $3.Literal},
            Type:  "",
        }
        $$ = append($1, arg)
    }
  | args_decl COMMA T PIDENTIFIER
  {
        arg := &ast.ArgDeclaration{
            Token: $4,
            Name:  &ast.Identifier{Token: $4, Value: $4.Literal},
            Type:  "T",
        }
        $$ = append($1, arg)
    }
  | PIDENTIFIER
  {
        $$ = []*ast.ArgDeclaration{
            &ast.ArgDeclaration{
                Token: $1,
                Name:  &ast.Identifier{Token: $1, Value: $1.Literal},
                Type:  "",
            },
        }
    }
  | T PIDENTIFIER
  {
        $$ = []*ast.ArgDeclaration{
            &ast.ArgDeclaration{
                Token: $2,
                Name:  &ast.Identifier{Token: $2, Value: $2.Literal},
                Type:  "T",
            },
        }
    }
  ;

args:
    args COMMA PIDENTIFIER
    {
        expr := &ast.Identifier{Token: $3, Value: $3.Literal}
        $$ = append($1, expr)
    }
  | PIDENTIFIER
   {
        $$ = []ast.Expression{
            &ast.Identifier{Token: $1, Value: $1.Literal},
        }
    }
  ;

expression:
    value
    {
        $$ = $1
    }
  | value PLUS value
  {
        $$ = &ast.BinaryExpression{
            Token:    $2,
            Left:     $1,
            Operator: "+",
            Right:    $3,
        }
    }
  | value MINUS value
  {
        $$ = &ast.BinaryExpression{
            Token:    $2,
            Left:     $1,
            Operator: "-",
            Right:    $3,
        }
    }
  | value MULT value
  {
        $$ = &ast.BinaryExpression{
            Token:    $2,
            Left:     $1,
            Operator: "*",
            Right:    $3,
        }
    }
  | value DIVIDE value
  {
        $$ = &ast.BinaryExpression{
            Token:    $2,
            Left:     $1,
            Operator: "/",
            Right:    $3,
        }
    }
  | value MODULO value
  {
        $$ = &ast.BinaryExpression{
            Token:    $2,
            Left:     $1,
            Operator: "%",
            Right:    $3,
        }
    }
  ;

condition:
    value EQUALS value
    {
        $$ = &ast.BinaryExpression{
            Token:    $2,
            Left:     $1,
            Operator: "==",
            Right:    $3,
        }
    }
  | value NEQUALS value
  {
        $$ = &ast.BinaryExpression{
            Token:    $2,
            Left:     $1,
            Operator: "!=",
            Right:    $3,
        }
    }
  | value GE value
  {
        $$ = &ast.BinaryExpression{
            Token:    $2,
            Left:     $1,
            Operator: ">",
            Right:    $3,
        }
  }
  | value LE value
  {
        $$ = &ast.BinaryExpression{
            Token:    $2,
            Left:     $1,
            Operator: "<",
            Right:    $3,
        }
    }
  | value GEQ value
  {
        $$ = &ast.BinaryExpression{
            Token:    $2,
            Left:     $1,
            Operator: ">=",
            Right:    $3,
        }
    }
  | value LEQ value
  {
        $$ = &ast.BinaryExpression{
            Token:    $2,
            Left:     $1,
            Operator: "<=",
            Right:    $3,
        }
    }
  ;

value:
    NUM
    {
        $$ = &ast.IntegerLiteral{
            Token: $1,
            Value: atoi($1.Literal),
        }
    }
  | identifier
  {
        $$ = $1
    }
  ;

identifier:
    PIDENTIFIER
    {
        $$ = &ast.Identifier{
            Token: $1,
            Value: $1.Literal,
        }
    }
  | PIDENTIFIER LBRACKET PIDENTIFIER RBRACKET
  {
        name := &ast.Identifier{Token: $1, Value: $1.Literal}
        index := &ast.Identifier{Token: $3, Value: $3.Literal}
        $$ = &ast.ArrayAccess{
            Token:     $2,
            ArrayName: name,
            Index:     index,
        }
    }
  | PIDENTIFIER LBRACKET NUM RBRACKET
  {
        name := &ast.Identifier{Token: $1, Value: $1.Literal}
        index := &ast.IntegerLiteral{Token: $3, Value: atoi($3.Literal)}
        $$ = &ast.ArrayAccess{
            Token:     $2,
            ArrayName: name,
            Index:     index,
        }
    }
  ;

%%