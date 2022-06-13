package util

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/linakesi/lnksutils"
	"io/ioutil"
)

func LoadOrCreatePrivateKey(pemPath string) (pcrypto.PrivKey, error) {
	if !lnksutils.IsFileExist(pemPath) {
		err := generatePrivateKeytoPEM(pemPath)
		if err != nil {
			return nil, err
		}
	}
	return LoadPrivateKey(pemPath)
}

func LoadPrivateKey(pemPath string) (pcrypto.PrivKey, error) {
	key, err := LoadPEMPrivateKey(pemPath)
	if err != nil {
		return nil, err
	}
	pk, _, err := pcrypto.KeyPairFromStdKey(key)
	return pk, err
}

func LoadPEMPrivateKey(where string) (crypto.PrivateKey, error) {
	privPEM, err := ioutil.ReadFile(where)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode([]byte(privPEM))
	if block == nil {
		return nil, errors.New("failed to parse Private Key")
	}

	switch block.Type {
	case "PRIVATE KEY":
		return x509.ParsePKCS8PrivateKey(block.Bytes)
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	default:
		return nil, fmt.Errorf("Not support Private Key type: %s", block.Type)
	}
}

func generatePrivateKeytoPEM(filePath string) error {
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	bs, err := x509.MarshalPKCS8PrivateKey(k)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	err = pem.Encode(buf, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: bs,
	})
	if err != nil {
		return err
	}

	return lnksutils.SaveToFile(buf, filePath,
		lnksutils.WithAtomicSave,
		lnksutils.WithFileMode(0400),
	)
}
