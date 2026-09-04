package constant

type Code int64

const (
	Success             Code = 0    // 成功
	Unauthorized        Code = 1001 // 认证失败 (token 无效/过期)
	Forbidden           Code = 1002 // 权限不足
	BadRequest          Code = 2001 // 请求参数错误
	NotFound            Code = 3001 // 资源不存在
	InternalServerError Code = 4001 // InternalServerError
)
