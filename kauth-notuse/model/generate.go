package authmodel

import (
	"crypto/sha256"
	"math/rand"
)

var safeCharset = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~")

// GenerateState 生成随机状态码，长度在 32 到 64 之间
func GenerateState() []byte {
	length := 32 + rand.Intn(33)
	b := make([]byte, length)
	for i := range b {
		b[i] = safeCharset[rand.Intn(len(safeCharset))]
	}
	return b
}

// GenerateCodeVerifier 生成固定长度为 43 的代码验证器
func GenerateCodeVerifier() []byte {
	length := 43
	b := make([]byte, length)
	for i := range b {
		b[i] = safeCharset[rand.Intn(len(safeCharset))]
	}
	return b
}

// GenerateCodeChallenge 生成代码挑战
func GenerateCodeChallenge(verifier []byte) []byte {
	h := sha256.Sum256(verifier)
	// 对哈希结果进行 Base64 URL 安全编码，这里直接处理以保证长度
	encoded := make([]byte, 0, len(h)*4/3)
	var i int
	for i = 0; i < len(h)-2; i += 3 {
		// 编码逻辑
		encoded = append(encoded, safeCharset[(h[i]>>2)&0x3f])
		encoded = append(encoded, safeCharset[((h[i]&0x3)<<4)|((h[i+1]>>4)&0xf)])
		encoded = append(encoded, safeCharset[((h[i+1]&0xf)<<2)|((h[i+2]>>6)&0x3)])
		encoded = append(encoded, safeCharset[h[i+2]&0x3f])
	}
	if i < len(h) {
		encoded = append(encoded, safeCharset[(h[i]>>2)&0x3f])
		if i+1 < len(h) {
			encoded = append(encoded, safeCharset[((h[i]&0x3)<<4)|((h[i+1]>>4)&0xf)])
			encoded = append(encoded, safeCharset[(h[i+1]&0xf)<<2])
		} else {
			encoded = append(encoded, safeCharset[(h[i]&0x3)<<4])
		}
	}
	// 截取前 43 个字符，保证长度一致
	if len(encoded) > 43 {
		encoded = encoded[:43]
	}
	return encoded
}
