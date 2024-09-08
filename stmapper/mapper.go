package stmapper

type MappingResult string

var result MappingResult = "implementation not generated, run stmapper"

func BuildMapping(src, dst interface{}) MappingResult {
	return result
}

func BuildMappingFrom(args ...interface{}) MappingResult {
	return result
}
