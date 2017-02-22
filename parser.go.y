%{
package golisp
%}

%union{
	s Value
	ss []Value
	tok token
}

%type<s> entry s s_quoted s_inner s_inner_term
%type<ss> ss_1 ss

%token<tok> SYM STR NUM
%token<tok> LPAREN RPAREN LBRACK RBRACK DOT TRUE FALSE QUOTE QUASIQUOTE UNQUOTE UNQUOTE_SPLICING
%token<tok> UNUSED

%%

entry
	: s
	{
		$$ = $1
		yylex.(*lexer).result = $$
	}
s
	: LPAREN RPAREN { $$ = Nil{} }
	| LBRACK RBRACK { $$ = Nil{} }
	| LPAREN s_inner RPAREN { $$ = $2 }
	| LBRACK s_inner RBRACK { $$ = $2 }
	| s_quoted { $$ = $1 }
	| NUM { $$ = Num{$1.num} }
	| SYM { $$ = Sym{$1.lit} }
	| STR { $$ = Str{$1.str} }
	| TRUE { $$ = Bool{true} }
	| FALSE { $$ = Bool{false} }
s_quoted
	: QUOTE s { $$ = Quote($2) }
	| QUASIQUOTE s { $$ = Quasiquote($2) }
	| UNQUOTE s { $$ = Unquote($2) }
	| UNQUOTE_SPLICING s { $$ = UnquoteSplicing($2) }
s_inner
	: ss_1 s_inner_term
	{
		$$ = $2
		for i := range $1 {
			$$ = Cons{$1[len($1)-1-i], $$}
		}
	}
s_inner_term
	: { $$ = Nil{} }
	| DOT s { $$ = $2 }
ss_1
	: ss s { $$ = append($1, $2) }
ss
	: { $$ = make([]Value, 0, 4) }
	| ss s { $$ = append($1, $2) }
