%{
    // Package declaration and imports
    package main
    import (
        "fmt"
    )
%}

%start program_all

%token PROCEDURE IS BEGIN END PROGRAM IF THEN ELSE ENDIF WHILE DO ENDWHILE REPEAT UNTIL
%token FOR FROM TO DOWNTO ENDFOR READ WRITE T
%token ASSIGN PLUS MINUS TIMES DIVIDE MOD EQUAL NEQUAL GREATER LESS GEQ LEQ
%token LPAREN RPAREN LBRACK RBRACK SEMICOLON COMMA COLON
%token NUM PIDENTIFIER

%%

program_all:
    procedures main
    ;

procedures:
    procedures PROCEDURE proc_head IS declarations BEGIN commands END
    | procedures PROCEDURE proc_head IS BEGIN commands END
    | /* empty */
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
    | FOR pidentifier FROM value TO value DO commands ENDFOR
    | FOR pidentifier FROM value DOWNTO value DO commands ENDFOR
    | proc_call SEMICOLON
    | READ identifier SEMICOLON
    | WRITE value SEMICOLON
    ;

proc_head:
    pidentifier LPAREN args_decl RPAREN
    ;

proc_call:
    pidentifier LPAREN args RPAREN
    ;

declarations:
    declarations COMMA pidentifier
    | declarations COMMA pidentifier LBRACK num COLON num RBRACK
    | pidentifier
    | pidentifier LBRACK num COLON num RBRACK
    ;

args_decl:
    args_decl COMMA pidentifier
    | args_decl COMMA T pidentifier
    | pidentifier
    | T pidentifier
    ;

args:
    args COMMA pidentifier
    | pidentifier
    ;

expression:
    value
    | value PLUS value
    | value MINUS value
    | value TIMES value
    | value DIVIDE value
    | value MOD value
    ;

condition:
    value EQUAL value
    | value NEQUAL value
    | value GREATER value
    | value LESS value
    | value GEQ value
    | value LEQ value
    ;

value:
    num
    | identifier
    ;

identifier:
    pidentifier
    | pidentifier LBRACK pidentifier RBRACK
    | pidentifier LBRACK num RBRACK
    ;

pidentifier:
    PIDENTIFIER
    ;

num:
    NUM
    ;

%%

/* Additional Go code for semantic actions or helper functions can be added here. */
