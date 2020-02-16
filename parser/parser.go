package parser

import (
	"io"
	"podojjit/types"
)

var _ Parser = (*ParserImpl)(nil)

type Parser interface {
	Parse(r io.Reader) *types.Program
}

type ParserImpl struct {
	program *types.Program
}

func (pi *ParserImpl) Parse(r io.Reader) *types.Program {
	var c byte
	var err error

	cmd := make([]byte, 1)

	for {
		_, err = r.Read(cmd)
		if err == io.EOF {
			break
		}

		c = cmd[0]
		if c == '>' || c == '<' || c == '+' || c == '-' || c == '.' ||
			c == ',' || c == '[' || c == ']' {
			pi.program.AddInstruction(c)
		}
	}
	return pi.program
}

func NewParser() Parser {
	return &ParserImpl{
		program: types.NewProgram(),
	}
}
