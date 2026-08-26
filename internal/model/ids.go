// Package model 定义濒危语言声调证据复核台的核心实体、状态枚举与领域错误。
// 所有 ID 使用标准库 crypto/rand 生成，避免引入额外依赖，保证离线可构建。
package model

import (
	"crypto/rand"
	"encoding/hex"
)

// NewID 生成带前缀的随机 ID（8 字节熵，16 进制 16 字符）。
// 前缀用于区分类别（batch/speaker/segment/...），空前缀则只返回随机串。
func NewID(prefix string) string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// 极不可能发生；退化为固定前缀，保证调用方不拿到空串。
		if prefix == "" {
			return "id-fallback"
		}
		return prefix + "-fallback"
	}
	if prefix == "" {
		return hex.EncodeToString(b)
	}
	return prefix + "-" + hex.EncodeToString(b)
}
