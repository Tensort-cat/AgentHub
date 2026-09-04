package tool_enum

type ToolStatus int8

type UserToolStatus int8

const (
	Disable ToolStatus = iota + 1
	Enable
)

const (
	Unsubscribed UserToolStatus = iota + 1
	Subscribed
)
