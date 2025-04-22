package encrypt

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"strings"

	"github.com/hkensame/goken/pkg/errors"
	"github.com/spaolacci/murmur3"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/sha3"
)

const (
	defaultSaltLen       = 16
	defaultIterations    = 100
	defaultKeyLen        = 32
	defaultHashAlgorithm = "sha512"
)

var (
	ErrUndefinedHashFunc = errors.New("未定义的HashFunc")
)

const (
	SHA1      = "sha1"
	SHA256    = "sha256"
	SHA512    = "sha512"
	SHA3_256  = "sha3-256"
	SHA3_512  = "sha3-512"
	MD5       = "md5"
	Murmur32  = "murmur32"
	Murmur64  = "murmur64"
	Murmur128 = "murmur128"
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
	salt, err1 := hex.DecodeString(parts[1])
	encoded, err2 := hex.DecodeString(parts[2])
	if err1 != nil || err2 != nil {
		return "", false
	}
	derived := pbkdf2.Key([]byte(raw), salt, defaultOptions.Iterations, defaultOptions.KeyLen, defaultOptions.HashFunction)
	if hmacEqual(derived, encoded) {
		return defaultHashAlgorithm, true
	}
	return "", false
}

func hmacEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var res byte
	for i := range a {
		res |= a[i] ^ b[i]
	}
	return res == 0
}

type Hasher struct {
	hashFunc hash.Hash
	salt     []byte
}

func NewHasher(hf hash.Hash, salt []byte) *Hasher {
	if hf == nil {
		hf = sha512.New()
	}
	return &Hasher{hashFunc: hf, salt: salt}
}

func (h *Hasher) Hash(data []byte) []byte {
	h.hashFunc.Reset()
	if len(h.salt) > 0 {
		h.hashFunc.Write(h.salt)
	}
	h.hashFunc.Write(data)
	return h.hashFunc.Sum(nil)
}

func (h *Hasher) Verify(data, exp []byte) bool {
	return bytes.Equal(exp, h.Hash(data))
}

func HashWithDefault(data []byte, hashKey string) []byte {
	hf, _ := NewHashFuncWithDefault(hashKey)
	return NewHasher(hf, nil).Hash(data)
}

func VerifyHashWithDefault(row []byte, enp []byte, hashKey string) bool {
	hf, _ := NewHashFuncWithDefault(hashKey)
	return bytes.Equal(NewHasher(hf, nil).Hash(row), enp)
}

func NewHashFuncWithDefault(algorithm string) (hash.Hash, string) {
	switch algorithm {
	case SHA1:
		return sha1.New(), algorithm
	case SHA256:
		return sha256.New(), algorithm
	case SHA512:
		return sha512.New(), algorithm
	case SHA3_256:
		return sha3.New256(), algorithm
	case SHA3_512:
		return sha3.New512(), algorithm
	case MD5:
		return md5.New(), algorithm
	case Murmur64:
		return murmur3.New64(), algorithm
	case Murmur128:
		return murmur3.New128(), algorithm
	case "", Murmur32:
		return murmur3.New32(), algorithm
	default:
		return sha512.New(), defaultHashAlgorithm
	}
}

func NewHashFunc(algorithm string) (hash.Hash, error) {
	switch algorithm {
	case SHA1:
		return sha1.New(), nil
	case SHA256:
		return sha256.New(), nil
	case SHA512:
		return sha512.New(), nil
	case SHA3_256:
		return sha3.New256(), nil
	case SHA3_512:
		return sha3.New512(), nil
	case MD5:
		return md5.New(), nil
	case Murmur64:
		return murmur3.New64(), nil
	case Murmur128:
		return murmur3.New128(), nil
	case "", Murmur32:
		return murmur3.New32(), nil
	default:
		return nil, ErrUndefinedHashFunc
	}
}

func NewHashFuncForPBKDF2(algorithm string) (func() hash.Hash, error) {
	switch algorithm {
	case SHA1:
		return sha1.New, nil
	case SHA256:
		return sha256.New, nil
	case SHA512:
		return sha512.New, nil
	case SHA3_256:
		return sha3.New256, nil
	case SHA3_512:
		return sha3.New512, nil
	default:
		return nil, ErrUndefinedHashFunc
	}
}

func NewHashFuncForPBKDF2WithDefault(algorithm string) (func() hash.Hash, string) {
	switch algorithm {
	case SHA1:
		return sha1.New, algorithm
	case SHA256:
		return sha256.New, algorithm
	case SHA512:
		return sha512.New, algorithm
	case SHA3_256:
		return sha3.New256, algorithm
	case SHA3_512:
		return sha3.New512, algorithm
	default:
		return sha512.New, defaultHashAlgorithm
	}
}
