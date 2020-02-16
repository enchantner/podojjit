package interp

import (
	"errors"
	"fmt"
	"podojjit/types"
)

var _ Interpreter = (*InterpreterImpl)(nil)

type Interpreter interface {
	Interpret(verbose bool) error
}

type InterpreterImpl struct {
	program *types.Program
	memory  int
	ops     []BFOp
}

// Optimizes a loop that starts at loop_start (the opening JUMP_IF_DATA_ZERO).
// The loop runs until the end of ops (implicitly there's a back-jump after the
// last op in ops).
//
// If optimization succeeds, returns a sequence of instructions that replace the
// loop; otherwise, returns an empty vector.
func (ii *InterpreterImpl) optimizeLoop(loopStart int) []BFOp {
	var newOps []BFOp
	var repeatedOp BFOp

	if len(ii.ops)-loopStart == 2 {
		repeatedOp = ii.ops[loopStart+1]
		if repeatedOp.kind == IncData || repeatedOp.kind == DecData {
			newOps = append(newOps, BFOp{LoopSetToZero, 0})
		} else if repeatedOp.kind == IncPtr {
			newOps = append(newOps, BFOp{LoopMovePtr, repeatedOp.argument})
		} else if repeatedOp.kind == DecPtr {
			newOps = append(newOps, BFOp{LoopMovePtr, -repeatedOp.argument})
		}
	} else if len(ii.ops)-loopStart == 5 {
		// Detect patterns: -<+> and ->+<
		if ii.ops[loopStart+1].kind == DecData &&
			ii.ops[loopStart+3].kind == IncData &&
			ii.ops[loopStart+1].argument == 1 &&
			ii.ops[loopStart+3].argument == 1 {
			if ii.ops[loopStart+2].kind == IncPtr &&
				ii.ops[loopStart+4].kind == DecPtr &&
				ii.ops[loopStart+2].argument == ii.ops[loopStart+4].argument {
				newOps = append(
					newOps,
					BFOp{LoopMoveData, ii.ops[loopStart+2].argument})
			} else if ii.ops[loopStart+2].kind == DecPtr &&
				ii.ops[loopStart+4].kind == IncPtr &&
				ii.ops[loopStart+2].argument == ii.ops[loopStart+4].argument {
				newOps = append(
					newOps,
					BFOp{LoopMoveData, -ii.ops[loopStart+2].argument})
			}
		}
	}
	return newOps
}

func (ii *InterpreterImpl) translate() error {
	var pc int
	var programSize = ii.program.Size()

	// Throughout the translation loop, this stack contains offsets (in the ops
	// vector) of open brackets (JUMP_IF_DATA_ZERO ops) waiting for a closing
	// bracket. Since brackets nest, these naturally form a stack. The
	// JUMP_IF_DATA_ZERO ops get added to ops with their argument set to 0 until a
	// matching closing bracket is encountered, at which point the argument can be
	// back-patched.
	var openBracketStack []int
	openBracketStackTop := -1
	var openBracketOffset int
	var instruction byte

	var start int
	var numRepeats int
	var kind Kind

	var optimizedLoop []BFOp

	for pc < programSize {
		instruction = ii.program.Instructions[pc]
		if instruction == '[' {
			// Place a jump op with a placeholder 0 offset. It will be patched-up to
			// the right offset when the matching ']' is found.
			openBracketStack = append(openBracketStack, len(ii.ops))
			openBracketStackTop++
			ii.ops = append(ii.ops, BFOp{kind: JumpIfDataZero, argument: 0})
			pc++
		} else if instruction == ']' {
			if len(openBracketStack) == 0 {
				return &RuntimeError{errors.New(
					fmt.Sprintf("unmatched closing ']' at pc=%d\n", pc),
				)}
			}
			openBracketOffset = openBracketStack[openBracketStackTop]
			openBracketStack = openBracketStack[:openBracketStackTop]
			openBracketStackTop--

			optimizedLoop = ii.optimizeLoop(openBracketOffset)

			if len(optimizedLoop) == 0 {
				// Now we have the offset of the matching '['. We can use it to create a
				// new jump op for the ']' we're handling, as well as patch up the offset
				// of the matching '['.
				ii.ops[openBracketOffset].argument = len(ii.ops)
				ii.ops = append(ii.ops, BFOp{kind: JumpIfDataNotZero, argument: openBracketOffset})
			} else {
				ii.ops = ii.ops[:openBracketOffset]
				ii.ops = append(ii.ops, optimizedLoop...)
			}
			pc++
		} else {
			// Not a jump; all the other ops can be repeated, so find where the repeat
			// ends.
			start = pc
			pc++
			for pc < programSize && ii.program.Instructions[pc] == instruction {
				pc++
			}
			// Here pc points to the first new instruction encountered, or to the end
			// of the program.
			numRepeats = pc - start

			// Figure out which op kind the instruction represents and add it to the
			// ops.
			kind = InvalidOp
			switch instruction {
			case '>':
				kind = IncPtr
			case '<':
				kind = DecPtr
			case '+':
				kind = IncData
			case '-':
				kind = DecData
			case ',':
				kind = ReadStdin
			case '.':
				kind = WriteStdout
			default:
				return &RuntimeError{errors.New(
					fmt.Sprintf("bad char '%b' at pc=%d\n", instruction, pc),
				)}
			}
			ii.ops = append(ii.ops, BFOp{kind, numRepeats})
		}
	}
	return nil
}

func (ii *InterpreterImpl) Interpret(verbose bool) error {
	err := ii.translate()
	if err != nil {
		return err
	}

	memory := make([]byte, ii.memory)

	pc := 0
	dataptr := 0

	var op BFOp
	var kind Kind

	for pc < len(ii.ops) {
		op = ii.ops[pc]
		kind = op.kind

		switch kind {
		case IncPtr:
			dataptr += op.argument
		case DecPtr:
			dataptr -= op.argument
		case IncData:
			memory[dataptr] = byte(int(memory[dataptr]) + op.argument)
		case DecData:
			memory[dataptr] = byte(int(memory[dataptr]) - op.argument)
		case ReadStdin:
			for i := 0; i < op.argument; i++ {
				fmt.Scanf("%d", memory[dataptr])
			}
		case WriteStdout:
			for i := 0; i < op.argument; i++ {
				fmt.Printf("%d", memory[dataptr])
			}
		case LoopSetToZero:
			memory[dataptr] = 0
		case LoopMovePtr:
			for memory[dataptr] > 0 {
				dataptr += op.argument
			}
		case LoopMoveData:
			if memory[dataptr] != 0 {
				memory[dataptr+op.argument] += memory[dataptr]
				memory[dataptr] = 0
			}
		case JumpIfDataZero:
			if memory[dataptr] == 0 {
				pc = op.argument
			}
		case JumpIfDataNotZero:
			if memory[dataptr] != 0 {
				pc = op.argument
			}
		case InvalidOp:
			return &RuntimeError{errors.New(
				fmt.Sprintf("INVALID_OP encountered on pc=%d", pc),
			)}
		}

		pc++
	}

	return nil
}

func NewInterpreter(p *types.Program, memory int) Interpreter {
	return &InterpreterImpl{program: p, memory: memory}
}
