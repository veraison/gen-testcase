// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"crypto"
	"crypto/rand"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"strings"
	"errors"
	"fmt"
	"reflect"

	cose "github.com/veraison/go-cose"
	"github.com/lestrrat-go/jwx/v2/jwk"
)


const (
	HeaderLabelMeta = int64(8)
	noAlg = cose.Algorithm(-65537)
)

func coseSignerFromJWK(j []byte) (cose.Signer, error) {
	alg, key, err := getAlgAndKeyFromJWK(j)
	if err != nil {
		return nil, err
	}

	return cose.NewSigner(alg, key)
}

func getAlgAndKeyFromJWK(j []byte) (cose.Algorithm, crypto.Signer, error) {
	var (
		err error
		k   jwk.Key
		crv elliptic.Curve
		alg cose.Algorithm
	)

	k, err = jwk.ParseKey(j)
	if err != nil {
		return noAlg, nil, err
	}

	var key crypto.Signer

	err = k.Raw(&key)
	if err != nil {
		return noAlg, nil, err
	}

	switch v := key.(type) {
	case *ecdsa.PrivateKey:
		alg = ellipticCurveToAlg(v.Curve)
		if alg == noAlg {
			return noAlg, nil, fmt.Errorf("unknown elliptic curve %v", crv)
		}
	case ed25519.PrivateKey:
		alg = cose.AlgorithmEd25519
	case *rsa.PrivateKey:
		alg = rsaJWKToAlg(k)
		if alg == noAlg {
			return noAlg, nil, fmt.Errorf("unknown RSA algorithm %q", k.Algorithm().String())
		}
	default:
		return noAlg, nil, fmt.Errorf("unknown private key type %v", reflect.TypeOf(key))
	}

	return alg, key, nil
}

func ellipticCurveToAlg(c elliptic.Curve) cose.Algorithm {
	switch c {
	case elliptic.P256():
		return cose.AlgorithmES256
	case elliptic.P384():
		return cose.AlgorithmES384
	case elliptic.P521():
		return cose.AlgorithmES512
	default:
		return noAlg
	}
}

func rsaJWKToAlg(k jwk.Key) cose.Algorithm {
	switch k.Algorithm().String() {
	case "PS256":
		return cose.AlgorithmPS256
	case "PS384":
		return cose.AlgorithmPS384
	case "PS512":
		return cose.AlgorithmPS512
	default:
		return noAlg
	}
}

func sign(
	payload []byte,
	meta []byte,
	contentType string,
	signer cose.Signer,
) ([]byte, error) {
	if signer == nil {
		return nil, errors.New("nil signer")
	}

	message := cose.NewSign1Message()

	message.Payload = payload

	alg := signer.Algorithm()

	if strings.Contains(alg.String(), "unknown algorithm value") {
		return nil, errors.New("signer has no algorithm")
	}

	message.Headers.Protected.SetAlgorithm(alg)
	message.Headers.Protected[cose.HeaderLabelContentType] = contentType
	if meta != nil {
		message.Headers.Protected[HeaderLabelMeta] = meta
	}

	err := message.Sign(rand.Reader, []byte(""), signer)
	if err != nil {
		return nil, fmt.Errorf("COSE Sign1 signature failed: %w", err)
	}

	wrap, err := message.MarshalCBOR()
	if err != nil {
		return nil, fmt.Errorf("signed-corim marshaling failed: %w", err)
	}

	return wrap, nil
}

