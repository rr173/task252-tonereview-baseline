package model

import "errors"

// 领域错误，供 HTTP 层映射为对应状态码。
var (
	// ErrNotFound 实体不存在（→404）。
	ErrNotFound = errors.New("not_found")
	// ErrBadRequest 请求语义非法（→400）。
	ErrBadRequest = errors.New("bad_request")
	// ErrConflict 并发/状态冲突（→409）。
	ErrConflict = errors.New("conflict")
	// ErrFrozen 冻结后不可修改（→409）。
	ErrFrozen = errors.New("frozen")
	// ErrSealed 批次已封存，拒绝一切修改（→409）。
	ErrSealed = errors.New("sealed")
	// ErrInvalidState 非法状态流转（→409）。
	ErrInvalidState = errors.New("invalid_state")
	// ErrDuplicate 幂等指纹冲突（→409）。
	ErrDuplicate = errors.New("duplicate")
)
