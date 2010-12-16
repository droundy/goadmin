package crypt

import (
	"io"
	"io/ioutil"
	"os"
	"hash"
	"encoding/binary"
	"compress/flate"
	"crypto/cipher"
	"crypto/aes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/rsa"
)

func CreateNewKey() (key string, e os.Error) {
	k := make([]byte, 32)
	_,e = rand.Read(k)
	return string(k), e
}

type hashReader struct {
	r io.ReadCloser
  h hash.Hash
	publickey *rsa.PublicKey
	togo int64
}
func (hr *hashReader) Read(data []byte) (read int, e os.Error) {
	if hr.togo <= 0 {
		return 0, os.NewError("Attempting to read past the end of the file...")
	}
	if int64(len(data)) > hr.togo {
		data = data[0:int(hr.togo)]
		read, e = io.ReadAtLeast(hr.r, data, int(hr.togo))
		hr.h.Write(data[0:read])
		if e != nil {
			return read, os.NewError("Error over-reading encrypted file: "+e.String())
		}
		e = os.NewError("Unexpected end of file in hashReader.")
	} else {
		read, e = io.ReadFull(hr.r, data)
		hr.h.Write(data)
		if e != nil {
			return read, os.NewError("Error reading encrypted file: "+e.String())
		}
	}
	hr.togo -= int64(read)
	if hr.togo == 0 {
		hsum := hr.h.Sum()
		csum := make([]byte, len(hsum))
		_, enew := io.ReadFull(hr.r, csum)
		if e != nil {
			return read, os.NewError("Error reading checksum: "+enew.String())
		}
		if string(hsum) != string(csum) {
			ioutil.WriteFile("/tmp/hsum", hsum, 0644)
			ioutil.WriteFile("/tmp/csum", csum, 0644)
			return read, os.NewError("Checksum mismatch!")
		}
		var siglen int32
		e = binary.Read(hr.r, binary.LittleEndian, &siglen)
		if e != nil {
			return read, os.NewError("Error reading length of signature: "+e.String())
		}
		sig := make([]byte, int(siglen))
		_, e = io.ReadFull(hr.r, sig)
		if e != nil {
			return read, os.NewError("Error reading signature: "+e.String())
		}
		e = verify(hr.publickey, hsum, sig)
		if e != nil { return read, e }
		hr.r.Close()
	}
	return
}

func Decrypt(key string, publickey PublicKey, rin io.Reader) (r io.Reader, length int64, sequence int64, e os.Error) {
	c, e := simpleCipher(key)
	if e != nil {
		e = os.NewError("Trouble reading the symmetric key: "+e.String())
		return
	}
	pub, e := readPublicKey(publickey)
	if e != nil {
		e = os.NewError("Trouble reading the public key: "+e.String())
		return
	}
	iv := make([]byte, c.BlockSize())
	_, e = io.ReadFull(rin, iv) // read the iv first (it's not encrypted)
	if e != nil {
		e = os.NewError("Trouble reading the iv: "+e.String())
		return
	}
	decrypter := cipher.NewCFBDecrypter(c, iv)
	rdec := flate.NewReader(cipher.StreamReader{decrypter, rin})
	e = binary.Read(rdec, binary.LittleEndian, &length)
	if e != nil {
		e = os.NewError("Trouble reading the file length: "+e.String())
		return
	}
	e = binary.Read(rdec, binary.LittleEndian, &sequence)
	if e != nil {
		e = os.NewError("Trouble reading the serial number: "+e.String())
		return
	}
	return &hashReader{rdec, sha256.New(), pub, length}, length, sequence, nil
}

type hashWriter struct {
  w io.WriteCloser
	cipher io.Writer
	h hash.Hash
	privatekey *rsa.PrivateKey
	togo int64
}
func (hw *hashWriter) Write(data []byte) (written int, e os.Error) {
	hw.h.Write(data)
	written, e = hw.w.Write(data)
	if e != nil { return }
	hw.togo -= int64(len(data))
	if hw.togo <= 0 {
		hash := hw.h.Sum()
		hw.w.Write(hash)
		sig, e := sign(hw.privatekey, hash)
		if e != nil { return 0, e }
		binary.Write(hw.w, binary.LittleEndian, int32(len(sig)))
		hw.w.Write(sig)
		hw.w.Close() // Push everything through to the cipher!
	}
	return
}

func Encrypt(key string, privatekey PrivateKey, win io.Writer, length, sequence int64) (w io.Writer, e os.Error) {
	c, e := simpleCipher(key)
	if e != nil { return }
	priv, e := readRSAKey(privatekey)
	if e != nil { return }
	iv := make([]byte, c.BlockSize())
	_,e = rand.Read(iv)
	if e != nil {
		return
	}
	_,e = win.Write(iv) // pass the iv across first
	if e != nil { return }
	wraw := cipher.StreamWriter{ cipher.NewCFBEncrypter(c, iv), win, nil }
	wenc := flate.NewWriter(wraw, flate.BestCompression)
	e = binary.Write(wenc, binary.LittleEndian, length)
	if e != nil { return }
	e = binary.Write(wenc, binary.LittleEndian, sequence)
	if e != nil { return }
	return &hashWriter{wenc, wraw, sha256.New(), priv, length}, nil
}

func simpleCipher(key string) (cipher.Block, os.Error) {
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
