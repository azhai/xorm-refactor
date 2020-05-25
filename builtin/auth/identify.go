package auth

import (
	"encoding/hex"
	"time"

	"github.com/azhai/gozzo-utils/cryptogy"
	"github.com/muyo/sno"
)

var saltPasswd ICipher

type ICipher interface {
	CreatePassword(plainText string) string
	VerifyPassword(plainText, cipherText string) bool
}

func Cipher() ICipher {
	if saltPasswd == nil { // 8位salt值，用$符号分隔开
		saltPasswd = cryptogy.NewSaltPassword(8, "$")
	}
	return saltPasswd
}

func NewSerialNo(n byte) string {
	return sno.New(n).String()
}

func NewTimeSerialNo(n byte, t time.Time) string {
	return sno.NewWithTime(n, t).String()
}

func NewToken(n byte) string {
	tail := cryptogy.RandSalt(2)
	token := sno.New(n).Bytes()
	token = append(token[:8], []byte(tail)...)
	return hex.EncodeToString(token)
}
