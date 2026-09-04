package model_enum

type ModelProvider int8

const (
	ModelProviderOpenAI ModelProvider = iota + 1
	ModelProviderArk
)
