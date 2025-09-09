package qbe

// Convention: temporaries of the form /%<name>.\d+/ and labels of the form
// /@<name>.\d+/ are reserved for automatic name generation.
import (
	"fmt"
	"strings"
)

type param struct {
	Type ABIType
	Name Temporary
}

// A Function represents a function definition in QBE IL.
type Function struct {
	Linkage                   // The linkage of the function, cannot be thread
	retType      RetType      // The return type of the function
	name         GlobalSymbol // The symbol that references the function
	env          *Temporary   // Parameter used to implement closures
	params       []param
	variadic     bool // Set if function is variadic.
	variadicFrom int  // Index where variadic parameters start (NEW FIELD)
	blocks       []*Block
	labelGen     uint
	tmpGen       uint
}

func (f *Function) isDefinition() {}

// newFunction returns a new [Function] with function name name, private linkage,
// return type retType, no env or any other parameters and it is not set as variadic.
func newFunction(name GlobalSymbol, retType RetType) *Function {
	if retType == nil {
		panic("return type cannot be nil")
	}
	return &Function{
		Linkage:      PrivateLinkage(),
		retType:      retType,
		name:         name,
		env:          nil,
		params:       nil,
		variadic:     false,
		variadicFrom: 0,
		blocks:       nil,
		labelGen:     0,
		tmpGen:       0,
	}
}

// Name returns the name of f.
func (f *Function) Name() GlobalSymbol {
	return f.name
}

// RetType returns the return type of f.
func (f *Function) RetType() RetType {
	return f.retType
}

// SetEnv sets the environment temporary of f to env.
func (f *Function) SetEnv(env Temporary) {
	f.env = &env
}

// SetVariadic sets f as variadic with variadic parameters starting from the current parameter count.
func (f *Function) SetVariadic() {
	f.variadic = true
	f.variadicFrom = len(f.params)
}

// SetVariadicFrom sets f as variadic with variadic parameters starting from the specified index.
func (f *Function) SetVariadicFrom(index int) {
	f.variadic = true
	f.variadicFrom = index
}

// InsertParam inserts at the end of the parameter list of f a new parameter named name with type type_.
func (f *Function) InsertParam(type_ ABIType, name Temporary) {
	if type_ == nil {
		panic("parameter type cannot be nil")
	}
	f.params = append(f.params, param{type_, name})
}

// InsertBlock inserts a new [Block] at the end of the function body,
// with label as the name of its entry point. Returns a reference to that block.
func (f *Function) InsertBlock(label Label) *Block {
	f.blocks = append(f.blocks, newBlock(label))
	return f.blocks[len(f.blocks)-1]
}

// InsertBlockAuto inserts a new [Block] at the end of the function body,
// with an auto-generated label of the form /@<name>\.\d+/. Refrain from creating
// labels of this form and using this function to ensure uniqueness.
// Returns a reference to that block.
func (f *Function) InsertBlockAuto(name string) *Block {
	label := Label(fmt.Sprintf("%v.%v", name, f.labelGen))
	f.labelGen++
	return f.InsertBlock(label)
}

// NewTemporary returns a new [Temporary] of the form /%<name>\.\d+/. Refrain
// from creating temporaries of this form and using this function to ensure uniqueness.
func (f *Function) NewTemporary(name string) Temporary {
	tmp := Temporary(fmt.Sprintf("%v.%v", name, f.tmpGen))
	f.tmpGen++
	return tmp
}

// String converts f to a string compatible with QBE code.
func (f *Function) String() string {
	builder := strings.Builder{}
	linkage := f.Linkage.String()
	if linkage != "" {
		builder.WriteString(linkage)
		builder.WriteByte(' ')
	}
	builder.WriteString("function ")
	if f.retType != VoidType() {
		builder.WriteString(f.retType.(ABIType).Name())
		builder.WriteByte(' ')
	}
	builder.WriteString(f.name.String())
	builder.WriteByte('(')
	if f.env != nil {
		builder.WriteString("env ")
		builder.WriteString(f.env.String())
		if len(f.params) > 0 || f.variadic {
			builder.WriteString(", ")
		}
	}

	// Write parameters with variadic marker at the correct position
	for i, param := range f.params {
		// Insert variadic marker before variadic parameters
		if f.variadic && i == f.variadicFrom {
			builder.WriteString("..., ")
		}

		builder.WriteString(param.Type.Name())
		builder.WriteByte(' ')
		builder.WriteString(param.Name.String())

		// Add comma if not the last parameter
		if i < len(f.params)-1 {
			builder.WriteString(", ")
		}
	}

	// Handle case where there are no variadic params but function is marked variadic
	if f.variadic && f.variadicFrom >= len(f.params) {
		if len(f.params) > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString("...")
	}

	builder.WriteString(") {\n")
	for _, block := range f.blocks {
		builder.WriteString(block.String())
	}
	builder.WriteString("}")
	return builder.String()
}
