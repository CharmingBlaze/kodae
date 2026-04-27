package check

// Kind of type
type Kind int

const (
	KInvalid Kind = iota
	KInt
	KFloat
	KStr
	KBool
	KVoid
	KNil // none literal
	KRange
	// KEnum: EnumName is set. Variants in separate Enum def.
	KEnum
	KStruct // .StructName and .StructDef set
	KOptional // .Opt is inner type; EnumName empty
	KByte   // 8-bit byte (C uint8)
	KPtr    // .Pointee is *Type (ptr[Pointee] in Clio)
	KList   // .Elem is the list element type (list[Elem] in Clio)
	// KResult: .Res is the success value type (result[Res] in Clio)
	KResult
)

// Type in checker
type Type struct {
	Kind       Kind
	EnumName   string // for KEnum, name of the enum
	EnumRef    *Enum  // for KEnum: shared enum definition
	StructName string
	StructDef  *Struct
	Opt        *Type // for KOptional
	// Pointee is the referent for KPtr; Elem for KList; Res for KResult.
	// These fields are used only for kinds that need them.
	Pointee   *Type
	Elem      *Type
	Res       *Type
	Variants  map[string]int // unused on Type; in Enum
}

// Struct is a file-scope user struct (fields resolved after parse).
type Struct struct {
	Name   string
	Order  []string
	Fields map[string]*Type
}

// Enum is a file-scope enum
type Enum struct {
	Name  string
	Index map[string]int
}

func (t *Type) String() string {
	if t == nil {
		return "?"
	}
	switch t.Kind {
	case KInt:
		return "int"
	case KFloat:
		return "float"
	case KStr:
		return "str"
	case KBool:
		return "bool"
	case KVoid:
		return "void"
	case KNil:
		return "none"
	case KRange:
		return "range"
	case KEnum:
		if t.EnumName != "" {
			return t.EnumName
		}
		return "enum"
	case KStruct:
		if t.StructName != "" {
			return t.StructName
		}
		return "struct"
	case KOptional:
		if t.Opt != nil {
			return t.Opt.String() + "?"
		}
		return "??"
	case KByte:
		return "byte"
	case KPtr:
		if t.Pointee != nil {
			return "ptr[" + t.Pointee.String() + "]"
		}
		return "ptr[?]"
	case KList:
		if t.Elem != nil {
			return "list[" + t.Elem.String() + "]"
		}
		return "list[?]"
	case KResult:
		if t.Res != nil {
			return "result[" + t.Res.String() + "]"
		}
		return "result[?]"
	default:
		return "?"
	}
}

func (a *Type) equal(b *Type) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.Kind != b.Kind {
		// int vs float coercion sites handle separately; not equal
		return false
	}
	switch a.Kind {
	case KEnum:
		if a.EnumRef == nil && b.EnumRef == nil {
			return a.EnumName == b.EnumName
		}
		if a.EnumRef == nil || b.EnumRef == nil {
			return a.EnumName == b.EnumName
		}
		return a.EnumRef == b.EnumRef
	case KStruct:
		if a.StructDef != nil && b.StructDef != nil {
			return a.StructDef == b.StructDef
		}
		return a.StructName == b.StructName
	case KOptional:
		return a.Opt.equal(b.Opt)
	case KPtr:
		if a.Pointee == nil && b.Pointee == nil {
			return true
		}
		if a.Pointee == nil || b.Pointee == nil {
			return false
		}
		return a.Pointee.equal(b.Pointee)
	case KByte:
		return true
	case KList:
		if a.Elem == nil && b.Elem == nil {
			return true
		}
		if a.Elem == nil || b.Elem == nil {
			return false
		}
		return a.Elem.equal(b.Elem)
	case KResult:
		if a.Res == nil && b.Res == nil {
			return true
		}
		if a.Res == nil || b.Res == nil {
			return false
		}
		return a.Res.equal(b.Res)
	}
	return true
}

// Builtin
var TpInt = &Type{Kind: KInt}
var TpFloat = &Type{Kind: KFloat}
var TpStr = &Type{Kind: KStr}
var TpBool = &Type{Kind: KBool}
var TpVoid = &Type{Kind: KVoid}
var TpNil = &Type{Kind: KNil}
var TpRange = &Type{Kind: KRange}
var TpByte  = &Type{Kind: KByte}

func optionalOf(inner *Type) *Type { return &Type{Kind: KOptional, Opt: inner} }

// isNumeric is true for int and float
func (t *Type) isNumeric() bool {
	if t == nil {
		return false
	}
	return t.Kind == KInt || t.Kind == KFloat
}

// coercesToString is true for values that can be coerced in str+T concatenation.
func coercesToString(t *Type) bool {
	if t == nil {
		return false
	}
	switch t.Kind {
	case KInt, KFloat, KBool, KEnum:
		return true
	default:
		return false
	}
}

// StructType returns a struct reference type.
func StructType(s *Struct) *Type {
	if s == nil {
		return nil
	}
	return &Type{Kind: KStruct, StructName: s.Name, StructDef: s}
}
