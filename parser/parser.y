%{
package parser

import "github.com/Meduza3/imp/token"

type YySymType struct {
	Token *token.Token
  yys int
}
%}

%token ILLEGAL
%token EOF

%token PROCEDURE IS BEGIN END PROGRAM
%token ASSIGN SEMICOLON
%token IF THEN ELSE ENDIF
%token WHILE DO ENDWHILE
%token REPEAT UNTIL
%token FOR FROM TO DOWNTO ENDFOR
%token READ WRITE
%token LPAREN RPAREN COMMA
%token LBRACKET RBRACKET COLON
%token T
%token PLUS MINUS MULT DIVIDE MODULO
%token EQUALS NEQUALS GE LE GEQ LEQ
%token NUM PIDENTIFIER

%start program_all

%%

program_all:
    procedures main
  ;

procedures:
    procedures PROCEDURE proc_head IS declarations BEGIN commands END
  | procedures PROCEDURE proc_head IS BEGIN commands END
  |
  ;

main:
    PROGRAM IS declarations BEGIN commands END
  | PROGRAM IS BEGIN commands END
  ;

commands:
    commands command
  | command
  ;

command:
    identifier ASSIGN expression SEMICOLON
  | IF condition THEN commands ELSE commands ENDIF
  | IF condition THEN commands ENDIF
  | WHILE condition DO commands ENDWHILE
  | REPEAT commands UNTIL condition SEMICOLON
  | FOR PIDENTIFIER FROM value TO value DO commands ENDFOR
  | FOR PIDENTIFIER FROM value DOWNTO value DO commands ENDFOR
  | proc_call SEMICOLON
  | READ identifier SEMICOLON
  | WRITE value SEMICOLON
  ;

proc_head:
    PIDENTIFIER LPAREN args_decl RPAREN
  ;

proc_call:
    PIDENTIFIER LPAREN args RPAREN
  ;

declarations:
    declarations COMMA PIDENTIFIER
  | declarations COMMA PIDENTIFIER LBRACKET NUM COLON NUM RBRACKET
  | PIDENTIFIER
  | PIDENTIFIER LBRACKET NUM COLON NUM RBRACKET
  ;

args_decl:
    args_decl COMMA PIDENTIFIER
  | args_decl COMMA T PIDENTIFIER
  | PIDENTIFIER
  | T PIDENTIFIER
  ;

args:
    args COMMA PIDENTIFIER
  | PIDENTIFIER
  ;

expression:
    value
  | value PLUS value
  | value MINUS value
  | value MULT value
  | value DIVIDE value
  | value MODULO value
  ;

condition:
    value EQUALS value
  | value NEQUALS value
  | value GE value
  | value LE value
  | value GEQ value
  | value LEQ value
  ;

value:
    NUM
  | identifier
  ;

identifier:
    PIDENTIFIER
  | PIDENTIFIER LBRACKET PIDENTIFIER RBRACKET
  | PIDENTIFIER LBRACKET NUM RBRACKET
  ;

%%