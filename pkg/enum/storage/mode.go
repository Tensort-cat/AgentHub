package storage_enum

/*
	三种存储管理模式：
	1. 云模式
	2. 本地模式
	3. 混合模式
*/

type StorageMode int8

const (
	Cloud StorageMode = iota + 1
	Local
	Hybrid
)
