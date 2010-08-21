package crypt

import (
	"io"
	"io/ioutil"
	"os"
	"hash"
	"encoding/binary"
	"crypto/block"
	"crypto/aes"
	"crypto/rand"
	"crypto/sha256"
)

func CreateNewKey() (key string, e os.Error) {
	k := make([]byte, 32)
	_,e = rand.Read(k)
	return string(k), e
}

type hashReader struct {
	r io.Reader
  h hash.Hash
	togo int64
}
func (hr *hashReader) Read(data []byte) (read int, e os.Error) {
	if hr.togo <= 0 {
		return 0, os.EOF
	}
	if int64(len(data)) > hr.togo {
		data = data[0:int(hr.togo)]
		read, e = io.ReadAtLeast(hr.r, data, int(hr.togo))
		hr.h.Write(data[0:read])
		if e != nil { return }
		e = os.EOF
	} else {
		read, e = io.ReadFull(hr.r, data)
		hr.h.Write(data)
		if e != nil { return }
	}
	hr.togo -= int64(read)
	if hr.togo == 0 {
		hsum := hr.h.Sum()
		csum := make([]byte, len(hsum))
		_, enew := io.ReadFull(hr.r, csum)
		if enew != nil { return read, enew }
		if string(hsum) != string(csum) {
			ioutil.WriteFile("/tmp/hsum", hsum, 0644)
			ioutil.WriteFile("/tmp/csum", csum, 0644)
			return read, os.NewError("Checksum mismatch!")
		}
	}
	return
}

func Decrypt(key string, rin io.Reader) (r io.Reader, length int64, e os.Error) {
	c, e := simpleCipher(key)
	iv := make([]byte, 16)
	_, e = io.ReadFull(rin, iv) // read the iv first (it's not encrypted)
	if e != nil {
		return
	}
	r = block.NewCBCDecrypter(c, iv, rin)
	e = binary.Read(r, binary.LittleEndian, &length)
	if e != nil { return }
	return &hashReader{r, sha256.New(), length}, length, nil
}

type hashWriter struct {
  w io.Writer
	h hash.Hash
	togo int64
}
func (hw *hashWriter) Write(data []byte) (written int, e os.Error) {
	hw.h.Write(data)
	written, e = hw.w.Write(data)
	if e != nil { return }
	hw.togo -= int64(len(data))
	if hw.togo <= 0 {
		hw.w.Write(hw.h.Sum())
		hw.w.Write([]byte("a bit more padding here..."))
	}
	return
}

func Encrypt(key string, win io.Writer, length int64) (w io.Writer, e os.Error) {
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
	binary.Write(w, binary.LittleEndian, length)
	return &hashWriter{w, sha256.New(), length}, nil
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
