package encrypt

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const (
	defaultSaltLen    = 16
	defaultIterations = 100
	defaultKeyLen     = 32
)

var defaultOptions = &EncryptOption{
	SaltLen:      defaultSaltLen,
	Iterations:   defaultIterations,
	KeyLen:       defaultKeyLen,
	HashFunction: sha512.New,
}

type EncryptOption struct {
	SaltLen      int
	Iterations   int
	KeyLen       int
	HashFunction func() hash.Hash
}

func generateSalt(length int) []byte {
	salt := make([]byte, length)
	rand.Read(salt)
	return salt
}

func Encrypt(raw string) string {
	return EncryptWithOption(raw, defaultOptions)
}

func EncryptWithOption(raw string, opt *EncryptOption) string {
	if raw == "" {
		return ""
	}
	salt := generateSalt(opt.SaltLen)
	derived := pbkdf2.Key([]byte(raw), salt, opt.Iterations, opt.KeyLen, opt.HashFunction)
	return fmt.Sprintf("pbkdf2$%x$%x", salt, derived)
}

func EncryptWithHash(raw, hashAlgo string) string {
	hashFunc, _ := NewHashFuncForPBKDF2WithDefault(hashAlgo)
	opt := &EncryptOption{
		SaltLen:      defaultSaltLen,
		Iterations:   defaultIterations,
		KeyLen:       defaultKeyLen,
		HashFunction: hashFunc,
	}
	return EncryptWithOption(raw, opt)
}

func Verify(raw, encrypted string) bool {
	_, ok := VerifyWithError(raw, encrypted)
	return ok
}

func VerifyWithError(raw, encrypted string) (string, bool) {
	if raw == "" || encrypted == "" {
		return "", false
	}

	parts := strings.SplitN(encrypted, "$", 3)
	if len(parts) != 3 || parts[0] != "pbkdf2" {
		return "", false
	}

	saltHex, encodedHex := parts[1], parts[2]
	salt, err1 := hex.DecodeString(saltHex)
	encoded, err2 := hex.DecodeString(encodedHex)
	if err1 != nil || err2 != nil {
		return "", false
	}

	buf := pbkdf2.Key([]byte(raw), salt, defaultOptions.Iterations, defaultOptions.KeyLen, defaultOptions.HashFunction)
	if hmacEqual(buf, encoded) {
		return "sha512", true
	}
	return "", false
}

// 防止时间攻击
func hmacEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var res byte
	for i := 0; i < len(a); i++ {
		res |= a[i] ^ b[i]
	}
	return res == 0
}

// type Encryptor struct {
// 	hashFunc func() hash.Hash
// }

// func NewEncryptor(hf func() hash.Hash) *Encryptor {
// 	return &Encryptor{hashFunc: hf}
// }

// func (e *Encryptor) Encrypt(obj any) (string, error) {
// 	e.hashFunc().
// }

// // Verify 验证加密字符串的签名，并反序列化原始对象
// func (e *Encryptor) Verify(str string, out any) error {
// 	parts := bytes.Split([]byte(str), []byte("."))
// 	if len(parts) != 2 {
// 		return errors.New("invalid format")
// 	}

// 	payload := string(parts[0])
// 	sig := string(parts[1])
// 	expectedSig := e.computeHash(payload)
// 	if sig != expectedSig {
// 		return errors.New("invalid signature")
// 	}

// 	raw, err := base64.RawURLEncoding.DecodeString(payload)
// 	if err != nil {
// 		return err
// 	}

// 	return json.Unmarshal(raw, out)
// }

// // computeHash 计算签名哈希
// func (e *Encryptor) computeHash(data string) string {
// 	h := e.hashFunc()
// 	io.WriteString(h, data)
// 	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
// }
