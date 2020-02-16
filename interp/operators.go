package interp

type Kind int

const (
	InvalidOp Kind = iota
	IncPtr
	DecPtr
	IncData
	DecData
	ReadStdin
	WriteStdout
	JumpIfDataZero
	JumpIfDataNotZero
	// optimized
	LoopSetToZero
	LoopMovePtr
	LoopMoveData
)

func OpNameByKind(kind Kind) byte {
	switch kind {
	case IncPtr:
		return '>'
	case DecPtr:
		return '<'
	case IncData:
		return '+'
	case DecData:
		return '-'
	case ReadStdin:
		return ','
	case WriteStdout:
		return '.'
	case JumpIfDataZero:
		return '['
	case JumpIfDataNotZero:
		return ']'
	case LoopSetToZero:
		return 's'
	case LoopMovePtr:
		return 'm'
	case LoopMoveData:
		return 'd'
	case InvalidOp:
		return 'x'
	}
	return 0
}

type BFOp struct {
	kind     Kind
	argument int
}
