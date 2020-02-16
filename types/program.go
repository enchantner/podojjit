package types

type Program struct {
	Instructions []byte
}

func (p *Program) AddInstruction(c byte) {
	p.Instructions = append(p.Instructions, c)
}

func (p *Program) Size() int {
	return len(p.Instructions)
}

func NewProgram() *Program {
	return &Program{}
}
