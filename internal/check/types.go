package check

// Kind of type
type Kind int

const (
	KInvalid Kind = iota
	KInt
	KFloat
	KF32 // C float in extern signatures only; widens to KFloat in Clio expressions
	KI32 // C int32_t in extern only; widens to KInt in expressions
	KU32 // C uint32_t in extern only; widens to KInt
	KU8  // C uint8_t in extern only; widens to KInt
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
	KAny // Dynamic/Any type (for JSON/Objects)
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
	// Pub / SrcFile: if !Pub, the struct type is only usable in SrcFile; if Pub, any file.
	Pub     bool
	SrcFile string
}

// Enum is a file-scope enum
type Enum struct {
	Name  string
	Index map[string]int
	Pub   bool
	File  string
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
	case KF32:
		return "f32"
	case KI32:
		return "i32"
	case KU32:
		return "u32"
	case KU8:
		return "u8"
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
	if a.Kind == KF32 || a.Kind == KI32 || a.Kind == KU32 || a.Kind == KU8 {
		return true
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
// TpF32 is C float, valid only in extern fn signatures; calls widen to TpFloat.
var TpF32 = &Type{Kind: KF32}
var TpI32 = &Type{Kind: KI32}
var TpU32 = &Type{Kind: KU32}
var TpU8 = &Type{Kind: KU8}
var TpStr = &Type{Kind: KStr}
var TpBool = &Type{Kind: KBool}
var TpVoid = &Type{Kind: KVoid}
var TpNil = &Type{Kind: KNil}
var TpRange = &Type{Kind: KRange}
var TpByte  = &Type{Kind: KByte}
var TpAny   = &Type{Kind: KAny}
var TpListStr = &Type{Kind: KList, Elem: TpStr}

func optionalOf(inner *Type) *Type { return &Type{Kind: KOptional, Opt: inner} }

// isNumeric is true for int, float, and f32
func (t *Type) isNumeric() bool {
	if t == nil {
		return false
	}
	return t.Kind == KInt || t.Kind == KFloat || t.Kind == KF32 || t.Kind == KI32 || t.Kind == KU32 || t.Kind == KU8
}

// coercesToString is true for values that can be coerced in str+T concatenation.
func coercesToString(t *Type) bool {
	if t == nil {
		return false
	}
	switch t.Kind {
	case KInt, KFloat, KBool, KEnum, KI32, KU32, KU8, KF32, KList, KStruct:
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
