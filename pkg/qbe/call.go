package qbe

import (
	"fmt"
	"strings"
)

type arg struct {
	Type ABIType
	Value
}

// A CallInst represents a call instruction.
type CallInst struct {
	dest         *dest      // Where the result of the call is stored, nil if no result.
	env          *Temporary // The env parameter, nil if not passed.
	variadic     bool       // Set to true if the function called is variadic. Defaults to false
	variadicFrom int        // Index where variadic arguments start (NEW FIELD)
	target       Value
	args         []arg
}

func (*CallInst) isInst() {}

// String converts inst to a string compatible with QBE code.
func (c *CallInst) String() string {
	var builder strings.Builder
	if c.dest != nil {
		builder.WriteString(c.dest.Name.String())
		builder.WriteString(" =")
		builder.WriteString(fmt.Sprint(c.dest.Type.Name()))
		builder.WriteByte(' ')
	}
	builder.WriteString("call ")
	builder.WriteString(fmt.Sprint(c.target))
	builder.WriteByte('(')
	if c.env != nil {
		builder.WriteString("env ")
		builder.WriteString(c.env.String())
		if len(c.args) > 0 || c.variadic {
			builder.WriteString(", ")
		}
	}

	// Write arguments with variadic marker at the correct position
	for i, arg := range c.args {
		// Insert variadic marker before variadic arguments
		if c.variadic && i == c.variadicFrom {
			builder.WriteString("..., ")
		}

		builder.WriteString(fmt.Sprint(arg.Type.Name()))
		builder.WriteByte(' ')
		builder.WriteString(fmt.Sprint(arg.Value))

		// Add comma if not the last argument
		if i < len(c.args)-1 {
			builder.WriteString(", ")
		}
	}

	// Handle case where there are no variadic args but function is marked variadic
	if c.variadic && c.variadicFrom >= len(c.args) {
		if len(c.args) > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString("...")
	}

	builder.WriteByte(')')
	return builder.String()
}

// InsertCall adds a call instruction to b with callee target and
// returns a pointer to the generated CallInst. The call instruction has no
// destination, no env variable and is not variadic.
func (b *Block) InsertCall(target Value) *CallInst {
	if target == nil {
		panic("target of call instruction cannot be nil")
	}
	inst := &CallInst{
		dest:         nil,
		env:          nil,
		variadic:     false,
		variadicFrom: 0,
		target:       target,
		args:         nil,
	}
	b.insertInstruction(inst)
	return b.insts[len(b.insts)-1].(*CallInst)
}

// SetDest stores the result of c to the temporary name with type type_.
func (c *CallInst) SetDest(name Temporary, type_ ABIType) {
	if type_ == nil {
		panic("call destination type cannot be nil")
	}
	c.dest = &dest{name, type_}
}

// SetEnv sets the environment temporary of c to env.
func (c *CallInst) SetEnv(env Temporary) {
	c.env = &env
}

// SetVariadic sets c as variadic with variadic arguments starting from the current argument count.
func (c *CallInst) SetVariadic() {
	c.variadic = true
	c.variadicFrom = len(c.args)
}

// SetVariadicFrom sets c as variadic with variadic arguments starting from the specified index.
func (c *CallInst) SetVariadicFrom(index int) {
	c.variadic = true
	c.variadicFrom = index
}

// InsertArg adds argument value with type type_ to the end of the argument list of c.
func (c *CallInst) InsertArg(type_ ABIType, value Value) {
	if type_ == nil {
		panic("call argument type cannot be nil")
	}
	if value == nil {
		panic("call argument value cannot be nil")
	}
	c.args = append(c.args, arg{type_, value})
}

// InsertVastart adds a vastart at the end of b with argument ap.
func (b *Block) InsertVastart(ap Value) {
	inst := newSimpleInstNoDest(vastart, ap)
	b.insertInstruction(inst)
}

// InsertVaarg adds a vaarg at the end of b with source ap and storing the argument to dest with type type_.
func (b *Block) InsertVaarg(dest Temporary, type_ BaseType, ap Value) {
	inst := newSimpleInst(vaarg, dest, type_, ap)
	b.insertInstruction(inst)
}
