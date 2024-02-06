package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"flag"
	"log"
	"os"
	"sort"

	"github.com/chkimes/image-attestation/internal"
	"github.com/google/go-tpm/legacy/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

var (
	tpmPath      = flag.String("tpm-path", "/dev/tpmrm0", "Device path for TPM")
	akLocation   = flag.Uint("ak-location", 0x81000003, "Location of AK public key")
	certLocation = flag.Uint("cert-location", 0x1c101d0, "Location of AK cert")

	bootMeasurementsLocation   = flag.String("boot-measurements", "/sys/kernel/security/tpm0/binary_bios_measurements", "File path for boot measurements")
	verityMeasurementsLocation = flag.String("verity-measurements", "/measurements/eventlog", "File path for verity measurements")

	outputPath = flag.String("output-path", "attest.json", "Attestation output location")
)

func main() {
	flag.Parse()

	rwc, err := tpm2.OpenTPM(*tpmPath)
	if err != nil {
		log.Fatalf("can't open TPM %s: %v", *tpmPath, err)
	}
	defer rwc.Close()

	akCertBytes, err := tpm2.NVRead(rwc, tpmutil.Handle(*certLocation))
	if err != nil {
		log.Fatalf("can't read AK cert at %x: %v", *certLocation, err)
	}

	akCert, err := x509.ParseCertificate(akCertBytes)
	if err != nil {
		log.Fatalf("can't parse AK cert: %v", err)
	}

	if akCert.PublicKeyAlgorithm.String() != "RSA" {
		log.Fatalf("TODO: currently only support RSA")
	}

	akPub := akCert.PublicKey.(*rsa.PublicKey)

	log.Printf("AK cert:")
	log.Printf("\tSubject: %s", akCert.Subject)
	log.Printf("\tPubkey Alg: %s", akCert.PublicKeyAlgorithm.String())
	log.Printf("\t\tModulus: %x", akPub.N)
	log.Printf("\t\tExponent: %d", akPub.E)
	log.Printf("\tPubkey: %s", akCert.PublicKeyAlgorithm)
	log.Printf("\tIssuer: %s", akCert.Issuer)
	log.Printf("\tSignature: %x", akCert.Signature)

	bootMeasurements, err := os.ReadFile(*bootMeasurementsLocation)
	if err != nil {
		log.Fatalf("couldn't read boot measurements: %v", err)
	}

	verityMeasurements, err := os.ReadFile(*verityMeasurementsLocation)
	if err != nil {
		log.Fatalf("couldn't read verity measurements: %v", err)
	}

	if false {
		log.Printf("%d %d", len(bootMeasurements), len(verityMeasurements))
	}

	nonce := make([]byte, 8)
	rand.Read(nonce)

	pcrsel := tpm2.PCRSelection{
		Hash: tpm2.AlgSHA256,
		PCRs: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11},
	}
	quoteData, quoteSig, err := tpm2.QuoteRaw(rwc, tpmutil.Handle(*akLocation), "", "", nonce, pcrsel, tpm2.AlgNull)
	if err != nil {
		log.Fatalf("couldn't quote PCRs: %v", err)
	}

	// PCR_Read only supports reading 8 PCRs at a time
	pcrsel = tpm2.PCRSelection{
		Hash: tpm2.AlgSHA256,
		PCRs: []int{0, 1, 2, 3, 4, 5, 6, 7},
	}
	pcrs, err := tpm2.ReadPCRs(rwc, pcrsel)
	if err != nil {
		log.Fatalf("couldn't read PCRs: %v", err)
	}

	// convert pcrs map to a list of PCRValue
	var pcrValues []internal.PCRValue
	for k, v := range pcrs {
		pcrValues = append(pcrValues, internal.PCRValue{
			Index: k,
			Value: v,
		})
	}

	// read remaining PCRs
	pcrsel = tpm2.PCRSelection{
		Hash: tpm2.AlgSHA256,
		PCRs: []int{8, 9, 11},
	}
	pcrs, err = tpm2.ReadPCRs(rwc, pcrsel)
	if err != nil {
		log.Fatalf("couldn't read PCRs: %v", err)
	}

	for k, v := range pcrs {
		pcrValues = append(pcrValues, internal.PCRValue{
			Index: k,
			Value: v,
		})
	}

	sort.Slice(pcrValues, func(i, j int) bool {
		return pcrValues[i].Index < pcrValues[j].Index
	})
	log.Printf("PCR Values:")
	for _, pcr := range pcrValues {
		log.Printf("\t%d: %x", pcr.Index, pcr.Value)
	}

	log.Printf("Quote Data: %x", quoteData)
	log.Printf("Quote Sig: %x", quoteSig)

	attestation := internal.Attestation{
		AkCert:         akCertBytes,
		BootEventLog:   bootMeasurements,
		VerityEventLog: verityMeasurements,
		QuoteData:      quoteData,
		QuoteSignature: quoteSig,
		PCRs:           pcrValues,
	}
	json, err := json.Marshal(attestation)
	if err != nil {
		log.Fatalf("couldn't serialize attestation: %v", err)
	}

	err = os.WriteFile(*outputPath, json, 0666)
	if err != nil {
		log.Fatalf("writing file: %v", err)
	}
}
