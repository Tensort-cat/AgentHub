package message_enum

type MessageType int8

const (
	MessageTypeUser MessageType = iota + 1
	MessageTypeAssistant
	MessageTypeTool
	MessageTypeSystem
)
