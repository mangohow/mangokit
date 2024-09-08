package internal

import (
	"testing"
)

func TestConvertInterface(t *testing.T) {
	t.Log(ConvertInterface().ConvertString("Name"))
	t.Log(ConvertInterface().ConvertString("Name"))
	t.Log(ConvertInterface().ConvertString("Name"))
	t.Log(ConvertInterface().ConvertString("Name"))
	t.Log(ConvertInterface().ConvertString("Name"))
	t.Log(ConvertInterface().ConvertString("Name"))
	t.Log(ConvertInterface().ConvertString("Name"))
}

func TestNumberConvertString(t *testing.T) {
	t.Log(NumberConvertString("int").ConvertString("Name"))
	t.Log(NumberConvertString("int8").ConvertString("Name"))
	t.Log(NumberConvertString("int16").ConvertString("Name"))
	t.Log(NumberConvertString("int32").ConvertString("Name"))
	t.Log(NumberConvertString("int64").ConvertString("Name"))
	t.Log(NumberConvertString("uint").ConvertString("Name"))
	t.Log(NumberConvertString("uint8").ConvertString("Name"))
	t.Log(NumberConvertString("uint16").ConvertString("Name"))
	t.Log(NumberConvertString("uint32").ConvertString("Name"))
	t.Log(NumberConvertString("uint64").ConvertString("Name"))
	t.Log(NumberConvertString("float32").ConvertString("Name"))
	t.Log(NumberConvertString("float64").ConvertString("Name"))
	t.Log(NumberConvertString("bool").ConvertString("Name"))
}

func TestIntConvertTime(t *testing.T) {
	t.Log(IntConvertTime("int", Unix).ConvertString("Name"))
	t.Log(IntConvertTime("int", UnixMilli).ConvertString("Name"))
	t.Log(IntConvertTime("int", UnixMicro).ConvertString("Name"))
	t.Log(IntConvertTime("int", UnixNano).ConvertString("Name"))
	t.Log(IntConvertTime("int64", Unix).ConvertString("Name"))
	t.Log(IntConvertTime("int64", UnixMilli).ConvertString("Name"))
	t.Log(IntConvertTime("int64", UnixMicro).ConvertString("Name"))
	t.Log(IntConvertTime("int64", UnixNano).ConvertString("Name"))
}

func TestInterfaceConvert(t *testing.T) {
	t.Log(InterfaceConvert("int").ConvertString("Name"))
	t.Log(InterfaceConvert("int32").ConvertString("Name"))
	t.Log(InterfaceConvert("int64").ConvertString("Name"))
	t.Log(InterfaceConvert("uint").ConvertString("Name"))
	t.Log(InterfaceConvert("int32").ConvertString("Name"))
	t.Log(InterfaceConvert("int64").ConvertString("Name"))
	t.Log(InterfaceConvert("string").ConvertString("Name"))
}

func TestNumberConvertEachOther(t *testing.T) {
	t.Log(NumberConvertEachOther("int").ConvertString("Name"))
	t.Log(NumberConvertEachOther("int32").ConvertString("Name"))
	t.Log(NumberConvertEachOther("int64").ConvertString("Name"))
	t.Log(NumberConvertEachOther("uint").ConvertString("Name"))
	t.Log(NumberConvertEachOther("uint32").ConvertString("Name"))
	t.Log(NumberConvertEachOther("uint64").ConvertString("Name"))
	t.Log(NumberConvertEachOther("float32").ConvertString("Name"))
	t.Log(NumberConvertEachOther("float64").ConvertString("Name"))
}

func TestStringConvertNumber(t *testing.T) {
	t.Log(StringConvertNumber("int").ConvertString("Name"))
	t.Log(StringConvertNumber("int8").ConvertString("Name"))
	t.Log(StringConvertNumber("int16").ConvertString("Name"))
	t.Log(StringConvertNumber("int32").ConvertString("Name"))
	t.Log(StringConvertNumber("int64").ConvertString("Name"))
	t.Log(StringConvertNumber("uint").ConvertString("Name"))
	t.Log(StringConvertNumber("uint8").ConvertString("Name"))
	t.Log(StringConvertNumber("uint16").ConvertString("Name"))
	t.Log(StringConvertNumber("uint32").ConvertString("Name"))
	t.Log(StringConvertNumber("uint64").ConvertString("Name"))
	t.Log(StringConvertNumber("float32").ConvertString("Name"))
	t.Log(StringConvertNumber("float64").ConvertString("Name"))
	t.Log(StringConvertNumber("bool").ConvertString("Name"))
}

func TestStringConvertTime(t *testing.T) {
	t.Log(StringConvertTime().ConvertString("Name"))
}

func TestTimeConvertInt(t *testing.T) {
	t.Log(TimeConvertInt("int", Unix).ConvertString("Name"))
	t.Log(TimeConvertInt("int", UnixMilli).ConvertString("Name"))
	t.Log(TimeConvertInt("int", UnixMicro).ConvertString("Name"))
	t.Log(TimeConvertInt("int", UnixNano).ConvertString("Name"))
	t.Log(TimeConvertInt("int64", Unix).ConvertString("Name"))
	t.Log(TimeConvertInt("int64", UnixMilli).ConvertString("Name"))
	t.Log(TimeConvertInt("int64", UnixMicro).ConvertString("Name"))
	t.Log(TimeConvertInt("int64", UnixNano).ConvertString("Name"))
}

func TestTimeConvertString(t *testing.T) {
	t.Log(TimeConvertString().ConvertString("Name"))
}
