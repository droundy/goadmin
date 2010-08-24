package crypt

import (
	"fmt"
	"big"
	"os"
	"crypto/rsa"
	"crypto/rand"
)

type PrivateKey string
type PublicKey string

func CreateRSAKeyPair() (public PublicKey, private PrivateKey, e os.Error) {
	// A 3072 bit key seems a reasonable length from what I've
	// read... for one thing, it's only needed to protect from an
	// attacker who has already compromised the 128 bit symmetric key,
	// either by brute-forcing it, or (more likely) by rooting the
	// client.  In addition, you aren't required to use the same private
	// key for all clients... in fact you'll have the strongest security
	// by going with the default of one private key per client.
	priv, e := rsa.GenerateKey(rand.Reader, 3072)
	if e != nil { return }
	return writePublicKey(priv.PublicKey), writeRSAKey(priv), nil
}

func writeRSAKey(key *rsa.PrivateKey) PrivateKey {
	return PrivateKey(fmt.Sprintf("N = %v\nE = %d\nD = %v\nP = %v\nQ = %v\n",
		key.N, key.E, key.D, key.P, key.Q))
}

func writePublicKey(key rsa.PublicKey) PublicKey {
	return PublicKey(fmt.Sprintf("N = %v\nE = %d\n", key.N, key.E))
}

func readRSAKey(x PrivateKey) (*rsa.PrivateKey, os.Error) {
	var k rsa.PrivateKey
	var Ns, Ds, Ps, Qs string
	_, e := fmt.Sscanf(string(x), "N = %s\nE = %d\nD = %s\nP = %s\nQ = %s\n", &Ns, &k.E, &Ds, &Ps, &Qs)
	if e != nil { return nil, os.NewError("Couldn't read file: "+e.String()) }
	var ok = true
	k.N, ok = big.NewInt(0).SetString(Ns, 0)
	if !ok { return nil, os.NewError("Unable to read k.N: "+Ns) }
	k.D, ok = big.NewInt(0).SetString(Ds, 0)
	if !ok { return nil, os.NewError("Unable to read k.D: "+Ds) }
	k.P, ok = big.NewInt(0).SetString(Ps, 0)
	if !ok { return nil, os.NewError("Unable to read k.P: "+Ps) }
	k.Q, ok = big.NewInt(0).SetString(Qs, 0)
	if !ok { return nil, os.NewError("Unable to read k.Q: "+Qs) }
	return &k, nil
}

func readPublicKey(x PublicKey) (*rsa.PublicKey, os.Error) {
	var k rsa.PublicKey
	var Ns string
	_, e := fmt.Sscanf(string(x), "N = %s\nE = %d\n", &Ns, &k.E)
	if e != nil { return nil, e }
	var ok = false
	k.N, ok = big.NewInt(0).SetString(Ns, 0)
	if !ok { return nil, os.NewError("Unable to read k.N: "+Ns) }
	return &k, nil
}

func sign(key *rsa.PrivateKey, sha256hash []byte) ([]byte, os.Error) {
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, rsa.HashSHA256, sha256hash)
	fmt.Fprintln(os.Stderr, "Length of sig is", len(sig))
	return sig, err
}

func verify(pub *rsa.PublicKey, sha256hash, sig []byte) os.Error {
	fmt.Fprintln(os.Stderr, "Length of sig is", len(sig))
	return rsa.VerifyPKCS1v15(pub, rsa.HashSHA256, sha256hash, sig)
}
