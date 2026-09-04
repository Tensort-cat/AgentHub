package tool_enum

type ToolType int8

const (
	ToolTypeMCP ToolType = iota + 1
	ToolTypeHTTP
	ToolTypeBuiltin
)
