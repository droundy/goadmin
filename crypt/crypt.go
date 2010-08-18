package crypt

import (
	"io"
	"os"
	"crypto/block"
	"crypto/aes"
	"crypto/rand"
)

func CreateNewKey() (key string, e os.Error) {
	k := make([]byte, 32)
	_,e = rand.Read(k)
	return string(k), e
}

func Decrypt(key string, rin io.Reader) (r io.Reader, e os.Error) {
	c, e := simpleCipher(key)
	iv := make([]byte, 16)
	n, e := rin.Read(iv) // read the iv first (it's not encrypted)
	if e != nil {
		return
	}
	if n != 16 {
		return r, os.NewError("Short read!");
	}
	r = block.NewCBCDecrypter(c, iv, rin)
	return
}

func Encrypt(key string, win io.Writer) (w io.Writer, e os.Error) {
	c, e := simpleCipher(key)
	if e != nil {
		return
	}
	iv := make([]byte, 16)
	_,e = rand.Read(iv)
	if e != nil {
		return
	}
	win.Write(iv) // pass the iv across first
	w = block.NewCBCEncrypter(c, iv, win)
	return
}

func simpleCipher(key string) (block.Cipher, os.Error) {
	var k []byte
	// pad the key (or truncate it) so it's the right size.
	if len(key) > 24 {
		k = make([]byte, 32)
	} else if len(key) > 16 {
		k = make([]byte, 24)
	} else {
		k = make([]byte, 16)
	}
	for i := range k {
		if i < len(key) {
			k[i] = key[i]
		}
	}
	return aes.NewCipher(k)
}
