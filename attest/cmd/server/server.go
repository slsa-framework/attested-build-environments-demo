package main

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/json"
	"encoding/pem"
	"flag"
	"log"
	"os"

	"github.com/chkimes/image-attestation/internal"
	"github.com/google/go-tpm/legacy/tpm2"
)

//go:embed azure-tl-intermediate.pem
var intermediateCA []byte

var (
	attestationPath = flag.String("attestation-path", "attest.json", "File path for the attestation document")
)

func main() {
	flag.Parse()

	attestationBytes, err := os.ReadFile(*attestationPath)
	if err != nil {
		log.Fatalf("couldn't read attestation: %v", err)
	}

	var attestation internal.Attestation
	err = json.Unmarshal(attestationBytes, &attestation)
	if err != nil {
		log.Fatalf("couldn't deserialize attestation: %v", err)
	}

	akCert, err := x509.ParseCertificate(attestation.AkCert)
	if err != nil {
		log.Fatalf("couldn't parse AK certificate: %v", err)
	}

	pemBlock, _ := pem.Decode(intermediateCA)
	intermediate, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		log.Fatalf("couldn't parse intermediate CA: %v", err)
	}

	roots := x509.NewCertPool()
	roots.AddCert(intermediate)

	// Validate the AK certificate from Azure VM vs the Azure vTPM intermediate CA
	_, err = akCert.Verify(x509.VerifyOptions{
		Roots:     roots,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	})

	if err != nil {
		log.Fatalf("couldn't verify AK certificate: %v", err)
	}

	buf := bytes.NewBuffer(attestation.QuoteSignature)
	sig, err := tpm2.DecodeSignature(buf)
	if err != nil {
		log.Fatalf("couldn't parse quote signature: %v", err)
	}

	if sig.Alg != tpm2.AlgRSASSA {
		log.Fatalf("only RSASSA is supported")
	}

	hash, err := sig.RSA.HashAlg.Hash()
	if err != nil {
		log.Fatalf("couldn't get hash algorithm: %v", err)
	}

	hasher := hash.New()
	hasher.Write(attestation.QuoteData)

	// Verify that the quote signature is valid and matches the pubkey in the AK certificate
	err = rsa.VerifyPKCS1v15(akCert.PublicKey.(*rsa.PublicKey), hash, hasher.Sum(nil), sig.RSA.Signature)
	if err != nil {
		log.Fatalf("quote signature verification failed: %v", err)
	}

	// TODO:
	// - Validate PCRs 0-9 against boot chain
	// - Validate PCR 11 against verity measurements
	// - Validate grub, kernel, initramfs hashes from boot event log
	// - Validate kernel command line to make sure it's not using break=
	// - Validate verity hash from verity event log
	// - Validate that verity and overlay were set up
}
