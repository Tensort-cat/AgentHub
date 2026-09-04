package model_enum

type ModelType int8

// 0 Chat 1 Embedding

const (
	CHAT ModelType = iota + 1
	EMBEDDING
)
