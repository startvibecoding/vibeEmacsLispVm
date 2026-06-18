package vm

// Parse parses all top-level expressions from src.
func Parse(src string) ([]Expr, error) {
	p := &parser{lexer: newLexer(src)}
	if err := p.advance(); err != nil {
		return nil, err
	}
	var exprs []Expr
	for p.cur.typ != tokenEOF {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)
	}
	return exprs, nil
}

type parser struct {
	lexer *lexer
	cur   token
}

func (p *parser) advance() error {
	tok, err := p.lexer.next()
	if err != nil {
		return err
	}
	p.cur = tok
	return nil
}

func (p *parser) parseExpr() (Expr, error) {
	switch p.cur.typ {
	case tokenEOF:
		return nil, newError(p.cur.pos, "unexpected EOF")
	case tokenLParen:
		return p.parseList()
	case tokenRParen:
		return nil, newError(p.cur.pos, "unexpected )")
	case tokenQuote:
		return p.parseQuote()
	case tokenString:
		v := String(p.cur.lit)
		return v, p.advance()
	case tokenAtom:
		v := parseAtom(p.cur.lit)
		return v, p.advance()
	default:
		return nil, newError(p.cur.pos, "unexpected token")
	}
}

func (p *parser) parseList() (Expr, error) {
	start := p.cur.pos
	if err := p.advance(); err != nil {
		return nil, err
	}
	items := List{}
	for p.cur.typ != tokenRParen {
		if p.cur.typ == tokenEOF {
			return nil, newError(start, "unterminated list")
		}
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		items = append(items, expr)
	}
	if err := p.advance(); err != nil {
		return nil, err
	}
	return items, nil
}

func (p *parser) parseQuote() (Expr, error) {
	start := p.cur.pos
	if err := p.advance(); err != nil {
		return nil, err
	}
	if p.cur.typ == tokenEOF {
		return nil, newError(start, "quote requires an expression")
	}
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return List{Symbol("quote"), expr}, nil
}
