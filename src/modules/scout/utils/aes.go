package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"math/rand"
	"time"
)

var AES string

// 随机生成AES
func RandomGenerateAES() {
	AESLen := 32
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	ByteKey := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < AESLen; i++ {
		result = append(result, ByteKey[r.Intn(len(ByteKey))])
	}
	AES = string(result)
}

// 对明文进行填充
func Padding(plainText []byte, blockSize int) []byte {
	// 计算要填充的长度
	n := blockSize - len(plainText)%blockSize

	// 对原来的明文填充n个n
	temp := bytes.Repeat([]byte{byte(n)}, n)
	plainText = append(plainText, temp...)
	return plainText
}

// 对密文删除填充
func UnPadding(cipherText []byte) ([]byte, error) {
	// 取出密文最后一个字节end
	end := cipherText[len(cipherText)-1]
	if int(end) > len(cipherText) {
		return nil, errors.New("invalid aes")
	}
	// 删除填充
	cipherText = cipherText[:len(cipherText)-int(end)]
	return cipherText, nil
}

// AEC加密（CBC模式）
func AES_CBC_Encrypt(plainText []byte, key string) string {
	// 指定加密算法，返回一个AES算法的Block接口对象
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic(err)
	}

	// 进行填充
	plainText = Padding(plainText, block.BlockSize())

	// 指定初始向量iv,长度和block的块尺寸一致
	iv := []byte("12345678abcdefgh")

	// 指定分组模式，返回一个BlockMode接口对象
	blockMode := cipher.NewCBCEncrypter(block, iv)

	// 加密连续数据库
	cipherText := make([]byte, len(plainText))
	blockMode.CryptBlocks(cipherText, plainText)

	// 返回密文
	return base64.StdEncoding.EncodeToString(cipherText)
}

// AEC解密（CBC模式）
func AES_CBC_Decrypt(cipherText string, key string) ([]byte, error) {
	decodeBytes, err := base64.StdEncoding.DecodeString(cipherText)

	// 指定解密算法，返回一个AES算法的Block接口对象
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	// 指定初始化向量IV,和加密的一致
	iv := []byte("12345678abcdefgh")

	// 指定分组模式，返回一个BlockMode接口对象
	blockMode := cipher.NewCBCDecrypter(block, iv)

	// 解密
	plainText := make([]byte, len(decodeBytes))
	blockMode.CryptBlocks(plainText, decodeBytes)

	// 删除填充
	plainText, err = UnPadding(plainText)
	if err != nil {
		return nil, err
	}
	return plainText, nil
}
