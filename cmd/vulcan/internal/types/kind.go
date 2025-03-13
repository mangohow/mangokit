package types

type TypeClass int

const (
	TypeInvalid TypeClass = iota
	TypeNumber
	TypeBool
	TypeString
	TypeStdTime
)

type Kind int

const (
	Invalid Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Float32
	Float64
	Interface
	Slice
	String
	StdTime
	Struct
)

func (k Kind) String() string {
	name, ok := kindNames[k]
	if !ok {
		return "Unknown"
	}
	return name
}

var (
	basicTypes = map[string]Kind{
		"bool":        Bool,
		"int":         Int,
		"int8":        Int8,
		"int16":       Int16,
		"int32":       Int32,
		"int64":       Int64,
		"uint":        Uint,
		"uint8":       Uint8,
		"uint16":      Uint16,
		"uint32":      Uint32,
		"uint64":      Uint64,
		"float32":     Float32,
		"float64":     Float64,
		"string":      String,
		"any":         Interface,
		"interface{}": Interface,
		"Time":        StdTime,
	}

	kindNames = map[Kind]string{
		Bool:      "bool",
		Int:       "int",
		Int8:      "int8",
		Int16:     "int16",
		Int32:     "int32",
		Int64:     "int64",
		Uint:      "uint",
		Uint8:     "uint8",
		Uint16:    "uint16",
		Uint32:    "uint32",
		Uint64:    "uint64",
		Float32:   "float32",
		Float64:   "float64",
		String:    "string",
		Interface: "interface{}",
		StdTime:   "Time",
	}
)

func IsBasicType(pkg, name string) bool {
	if name == "Time" && pkg == "time" {
		return true
	}
	_, ok := basicTypes[name]

	return ok
}

func GetKind(pkg, name string) Kind {
	v, ok := basicTypes[name]
	if v == StdTime {
		if pkg == "time" {
			return StdTime
		} else {
			return Invalid
		}
	}
	if !ok {
		return Invalid
	}
	return v
}

func ToTypeClass(kind Kind) TypeClass {
	switch {
	case IsNumber(kind):
		return TypeNumber
	case IsBool(kind):
		return TypeBool
	case IsString(kind):
		return TypeString
	case IsStdTime(kind):
		return TypeStdTime
	}

	return TypeInvalid
}

func IsBasicKind(kind Kind) bool {
	return kind == Bool ||
		kind == Int ||
		kind == Int8 ||
		kind == Int16 ||
		kind == Int32 ||
		kind == Int64 ||
		kind == Uint ||
		kind == Uint8 ||
		kind == Uint16 ||
		kind == Uint32 ||
		kind == Uint64 ||
		kind == Float32 ||
		kind == Float64 ||
		kind == String
}

func IsInt(kind Kind) bool {
	return kind == Int ||
		kind == Int8 ||
		kind == Int16 ||
		kind == Int32 ||
		kind == Int64
}

func IsUint(kind Kind) bool {
	return kind == Uint ||
		kind == Uint8 ||
		kind == Uint16 ||
		kind == Uint32 ||
		kind == Uint64
}

func IsFloat(kind Kind) bool {
	return kind == Float32 || kind == Float64
}

func IsNumber(kind Kind) bool {
	return kind == Int ||
		kind == Int8 ||
		kind == Int16 ||
		kind == Int32 ||
		kind == Int64 ||
		kind == Uint ||
		kind == Uint8 ||
		kind == Uint16 ||
		kind == Uint32 ||
		kind == Uint64 ||
		kind == Float32 ||
		kind == Float64
}

func IsString(kind Kind) bool {
	return kind == String
}

func IsBool(kind Kind) bool {
	return kind == Bool
}

func IsStdTime(kind Kind) bool {
	return kind == StdTime
}
