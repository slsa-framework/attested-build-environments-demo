package main

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/chkimes/image-attestation/internal"
	"github.com/google/go-tpm/legacy/tpm2"
	"golang.org/x/exp/slices"
)

//go:embed azure-tl-intermediate.pem
var intermediateCA []byte

var (
	attestationPath  = flag.String("attestation-path", "attest.json", "File path for the attestation document")
	kernelHash       = flag.String("kernel-hash", "dc13e62d8601fe4934edce87eee853f36904b7497adadb763e7e5ac7c096233d", "Expected kernel hash")
	initramfsHash    = flag.String("initramfs-hash", "b97ea6cc8b8668e49f5d1f92c0a921338f32788b289711e58f506c5179c462b8", "Expected initramfs hash")
	verityRootHash   = flag.String("verity-root-hash", "7c4770215babcd808f0b5d440bec40f1d0757fd25ca584a10781a00b7e239a0c", "Root hash for the verity device")
	expectedPcrsPath = flag.String("expected-pcrs-path", "expected-pcrs.json", "File path for the expected PCR values")
)

func main() {
	flag.Parse()

	verityRootHash, err := hex.DecodeString(*verityRootHash)
	if err != nil {
		log.Fatalf("couldn't decode verity root hash: %v", err)
	}

	attestationBytes, err := os.ReadFile(*attestationPath)
	if err != nil {
		log.Fatalf("couldn't read attestation: %v", err)
	}

	var attestation internal.Attestation
	err = json.Unmarshal(attestationBytes, &attestation)
	if err != nil {
		log.Fatalf("couldn't deserialize attestation: %v", err)
	}

	expectedPcrsBytes, err := os.ReadFile(*expectedPcrsPath)
	if err != nil {
		log.Fatalf("couldn't read expected PCR values: %v", err)
	}

	var expectedPcrs internal.ExpectedPCRs
	err = json.Unmarshal(expectedPcrsBytes, &expectedPcrs)
	if err != nil {
		log.Fatalf("couldn't deserialize expected PCR values: %v", err)
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

	quote, err := tpm2.DecodeAttestationData(attestation.QuoteData)
	if err != nil {
		log.Fatalf("couldn't parse quote: %v", err)
	}

	if quote.Type != tpm2.TagAttestQuote {
		log.Fatalf("attested data type is not a quote")
	}

	log.Printf("Nonce: %x", quote.ExtraData)

	// Validate that the PCRs in the quote match our expected PCRs of 0-9, 11
	PCRsCopy := make([]int, len(quote.AttestedQuoteInfo.PCRSelection.PCRs))
	copy(PCRsCopy, quote.AttestedQuoteInfo.PCRSelection.PCRs)
	sort.Slice(PCRsCopy, func(i, j int) bool {
		return PCRsCopy[i] < PCRsCopy[j]
	})

	if !reflect.DeepEqual(PCRsCopy, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11}) {
		log.Fatalf("unexpected PCRs (expected 0-9, 11): %v", PCRsCopy)
	}

	// Validate that the PCR values in the quote match the attestation document
	PCRValuesCopy := make([]internal.PCRValue, len(attestation.PCRs))
	copy(PCRValuesCopy, attestation.PCRs)
	sort.Slice(PCRValuesCopy, func(i, j int) bool {
		return PCRValuesCopy[i].Index < PCRValuesCopy[j].Index
	})

	hash, err = quote.AttestedQuoteInfo.PCRSelection.Hash.Hash()
	if err != nil {
		log.Fatalf("couldn't get PCR hash algorithm: %v", err)
	}
	hasher = hash.New()
	for _, pcr := range PCRValuesCopy {
		hasher.Write(pcr.Value)
	}
	pcrHash := hasher.Sum(nil)

	if !reflect.DeepEqual(pcrHash, []byte(quote.AttestedQuoteInfo.PCRDigest)) {
		log.Printf("PCR digest mismatch!")
		log.Printf("\tcalculated: %x", pcrHash)
		log.Fatalf("\tquoted:     %x", quote.AttestedQuoteInfo.PCRDigest)
	}

	log.Printf("PCR digest: %x", pcrHash)

	// At this point we know that:
	//    - The AK certificate is valid
	//    - The AK key was used to sign the quote
	//    - The quote is a valid TPM quote
	//    - The PCRs in the quote match the attestation document
	//    - The PCR indices in the quote and attestation match our expectations
	//
	// All the crypto shenanigans are now done, and we can start to validate the
	// contents of the event logs.

	idx := slices.IndexFunc(attestation.PCRs, func(pcr internal.PCRValue) bool {
		return pcr.Index == 11
	})
	if idx == -1 {
		log.Fatalf("no PCR 11 value found")
	}

	verityHash, err := validateVerityEventLog(attestation.VerityEventLog, attestation.PCRs[idx], hash)
	if err != nil {
		log.Fatalf("verity event log validation failed: %v", err)
	}

	if !bytes.Equal(verityHash, verityRootHash) {
		log.Fatalf("verity hash mismatch, expected %x, got %x", verityRootHash, verityHash)
	}

	log.Printf("verity hash: %x", verityHash)

	attestationPcrs := make(map[int][]byte)
	for _, pcr := range attestation.PCRs {
		attestationPcrs[pcr.Index] = pcr.Value
	}

	for _, expectedPcr := range expectedPcrs.PCRs {
		if attestedPcr, ok := attestationPcrs[expectedPcr.Index]; !ok {
			log.Fatalf("PCR %d missing from attestation", expectedPcr.Index)
		} else if !bytes.Equal(expectedPcr.Value, attestedPcr) {
			log.Fatalf("PCR %d value mismatch", expectedPcr.Index)
		}
	}

	log.Printf("Attestation verified successfully")

	// TODO:
	//   - Validate grub, kernel, initramfs hashes from boot event log
	//   - Validate kernel command line to make sure it's not using break=
	// ^-- These are optional since we are validating the PCRs exactly, but
	//     they are good to have for extra validation.
}

func validateVerityEventLog(verityLog []byte, pcrValue internal.PCRValue, hash crypto.Hash) ([]byte, error) {
	verityString := string(verityLog[:])
	verityLogs := strings.Split(verityString, "\n")
	verityLogs = slices.DeleteFunc(verityLogs, func(s string) bool {
		return s == ""
	})

	if len(verityLogs) != 4 {
		return nil, fmt.Errorf("unexpected number of verity logs: %d", len(verityLogs))
	}

	if verityLogs[0] != "VERITY_INITRAMFS" {
		return nil, fmt.Errorf("unexpected verity log: %s", verityLogs[0])
	}

	if verityLogs[2] != "VERITY_SUCCESS" {
		return nil, fmt.Errorf("unexpected verity log: %s", verityLogs[2])
	}

	if verityLogs[3] != "OVERLAY_SUCCESS" {
		return nil, fmt.Errorf("unexpected verity log: %s", verityLogs[3])
	}

	if strings.Index(verityLogs[1], "VERITY_HASH: ") != 0 {
		return nil, fmt.Errorf("unexpected verity log: %s", verityLogs[1])
	}

	verityHashHex := verityLogs[1][len("VERITY_HASH: "):]
	verityHash, err := hex.DecodeString(verityHashHex)
	if err != nil {
		return nil, fmt.Errorf("couldn't decode verity hash: %v", err)
	}

	pcr11 := newPCRHashValue(hash)
	for _, log := range verityLogs {
		err = pcr11.Extend([]byte(log))
		if err != nil {
			return nil, fmt.Errorf("couldn't extend PCR 11: %v", err)
		}
	}

	if !bytes.Equal(pcr11.value, pcrValue.Value) {
		return nil, fmt.Errorf("PCR 11 value mismatch, expected %x, got %x", pcr11.value, pcrValue.Value)
	}

	return verityHash, nil
}

type PCRHashValue struct {
	value []byte
	hash  crypto.Hash
}

func newPCRHashValue(hash crypto.Hash) PCRHashValue {
	return PCRHashValue{value: make([]byte, hash.Size()), hash: hash}
}

func (p *PCRHashValue) Extend(extension []byte) error {
	if len(p.value) != p.hash.Size() {
		return fmt.Errorf("invalid PCR value length, expected %d, got %d", p.hash.Size(), len(p.value))
	}

	hasher := p.hash.New()
	hasher.Write(extension)
	extension = hasher.Sum(nil)

	hasher.Reset()
	hasher.Write(p.value)
	hasher.Write(extension)
	p.value = hasher.Sum(nil)
	return nil
}
