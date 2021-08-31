package util

import (
	"github.com/tredoe/osutil/user/crypt/common"
	"github.com/tredoe/osutil/user/crypt/sha512_crypt"
)

func GetEncryptedPasswd(key string) (string, error) {
	c := sha512_crypt.New()
	salt := common.Salt{}
	saltBytes := salt.Generate(16)
	return c.Generate([]byte(key), saltBytes)
}
