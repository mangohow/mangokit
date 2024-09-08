package internal

import (
	"fmt"
	"strings"
)

type ConvertDef struct {
	Format       string
	Package      string
	OuterDeclare bool
}

type TimeAccuracy string

const (
	Unix      TimeAccuracy = "Unix"
	UnixMilli              = "UnixMilli"
	UnixMicro              = "UnixMicro"
	UnixNano               = "UnixNano"
)

func (cd ConvertDef) ConvertString(name string) string {
	if !cd.OuterDeclare {
		return fmt.Sprintf(cd.Format, name)
	}
	return fmt.Sprintf(cd.Format, strings.ToLower(name), name)
}

// StringConvertNumber string -> int*, uint*, bool
func StringConvertNumber(dst string) ConvertDef {
	cd := ConvertDef{
		Package:      "strconv",
		OuterDeclare: true,
	}

	if dst == "bool" {
		cd.Format = "%s, _ = strconv.ParseBool(%s)"
		return cd
	}

	format := `@, _ = strconv.ParseInt(@, 10, %s)`
	if strings.HasPrefix(dst, "uint") {
		format = `@, _ = strconv.ParseUint(@, 10, %s)`
	} else if strings.HasPrefix(dst, "float") {
		format = "@, _ = strconv.ParseFloat(@, %s)"
	}

	switch dst {
	case "int", "uint", "float64":
		cd.Format = fmt.Sprintf(format, "64")
	case "int8", "int16", "int32", "int64":
		cd.Format = fmt.Sprintf(format, strings.TrimPrefix(dst, "int"))
	case "uint8", "uint16", "uint32", "uint64":
		cd.Format = fmt.Sprintf(format, strings.TrimPrefix(dst, "uint"))
	case "float32":
		cd.Format = fmt.Sprintf(format, strings.TrimPrefix(dst, "float"))
	}

	cd.Format = strings.ReplaceAll(cd.Format, "@", "%s")
	return cd
}

// NumberConvertString int*, uint*, float, bool -> string
func NumberConvertString(dst string) (cd ConvertDef) {
	cd.Package = "strconv"
	switch {
	case dst == "bool":
		cd.Format = "strconv.FormatBool(%s)"
	case dst == "int64":
		cd.Format = "strconv.FormatInt(%s, 10)"
	case dst == "uint64":
		cd.Format = "strconv.FormatUint(%s, 10)"
	case dst == "float32":
		cd.Format = "strconv.FormatFloat(%s, 'g', -1, 32)"
	case dst == "float64":
		cd.Format = "strconv.FormatFloat(%s, 'g', -1, 64)"
	case strings.HasPrefix(dst, "int"):
		cd.Format = "strconv.FormatInt(int64(%s), 10)"
	case strings.HasPrefix(dst, "uint"):
		cd.Format = "strconv.FormatInt(uint64(%s), 10)"
	}

	return
}

// NumberConvertEachOther int*, uint*, float* -> int*, uint*, float*
func NumberConvertEachOther(dst string) (cd ConvertDef) {
	cd.Format = dst + "(%s)"
	return
}

// ConvertInterface all types -> interface{}
func ConvertInterface() (cd ConvertDef) {
	cd.Format = "%s"
	return
}

// InterfaceConvert interface{} t -> all types
func InterfaceConvert(dst string) (cd ConvertDef) {
	cd.Format = "%s.(" + dst + ")"
	return cd
}

// TimeConvertString time.Time -> string
func TimeConvertString() (cd ConvertDef) {
	cd.Package = "time"
	cd.Format = "%s.Format(time.DateTime)"
	return
}

// TimeConvertInt time.Time -> int, int64
func TimeConvertInt(dst string, ta TimeAccuracy) (cd ConvertDef) {
	cd.Package = "time"
	cd.Format = "%s." + string(ta) + "()"
	if dst == "int" {
		cd.Format = "int(" + cd.Format + ")"
	}

	return
}

// StringConvertTime string -> time.Time
func StringConvertTime() ConvertDef {
	cd := ConvertDef{
		Format:       "%s, _ = time.Parse(time.DateTime, %s)",
		Package:      "time",
		OuterDeclare: true,
	}

	return cd
}

// IntConvertTime int, int64 -> time.Time
func IntConvertTime(dst string, ta TimeAccuracy) (cd ConvertDef) {
	cd.Package = "time"
	switch ta {
	case Unix:
		cd.Format = "time.Unix(%s@%s, 0)"
	case UnixNano:
		cd.Format = "time.UnixMicro(%s@%s)"
	case UnixMilli, UnixMicro:
		cd.Format = "time." + string(ta) + "(%s@%s)"
	}
	if dst == "int" {
		cd.Format = fmt.Sprintf(cd.Format, "int64(", ")")
	} else {
		cd.Format = fmt.Sprintf(cd.Format, "", ")")
	}
	cd.Format = strings.ReplaceAll(cd.Format, "@", "%s")

	return cd
}
