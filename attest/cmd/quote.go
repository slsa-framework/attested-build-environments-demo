package cmd

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/chkimes/image-attestation/internal"
	"github.com/google/go-tpm/legacy/tpm2"
	"github.com/google/go-tpm/tpmutil"
	"github.com/spf13/cobra"
)

var quoteCmd = &cobra.Command{
	Use:   "quote",
	Short: "Get a TPM quote and output as a JSON-formatted attestation",
	RunE:  getQuote,
}

var (
	tpmPath                    string
	akLocation                 uint32
	certLocation               uint32
	bootMeasurementsLocation   string
	verityMeasurementsLocation string
	outputPath                 string
)

func init() {
	quoteCmd.Flags().StringVarP(
		&tpmPath,
		"tpm-path",
		"t",
		"/dev/tpmrm0",
		"Device path for TPM",
	)

	quoteCmd.Flags().Uint32VarP(
		&akLocation,
		"ak-location",
		"a",
		0x81000003,
		"Location of AK public key",
	)

	quoteCmd.Flags().Uint32VarP(
		&certLocation,
		"cert-location",
		"c",
		0x1c101d0,
		"Location of AK cert",
	)

	quoteCmd.Flags().StringVarP(
		&bootMeasurementsLocation,
		"boot-measurements",
		"b",
		"/sys/kernel/security/tpm0/binary_bios_measurements",
		"File path for boot measurements",
	)

	quoteCmd.Flags().StringVarP(
		&verityMeasurementsLocation,
		"verity-measurements",
		"v",
		"/measurements/eventlog",
		"File path for verity measurements",
	)

	quoteCmd.Flags().BoolVarP(
		&debugLogging,
		"debug",
		"d",
		false,
		"Flag enabling debug logging. Default: false",
	)

	quoteCmd.Flags().StringVarP(
		&outputPath,
		"output-path",
		"o",
		"attestion.json",
		"Attestation output location",
	)
}

func getQuote(_ *cobra.Command, args []string) error {

	// Access the TPM and its metadata
	rwc, err := tpm2.OpenTPM(tpmPath)
	if err != nil {
		return fmt.Errorf("can't open TPM %s: %w", tpmPath, err)
	}
	defer rwc.Close()

	akCertBytes, err := tpm2.NVRead(rwc, tpmutil.Handle(certLocation))
	if err != nil {
		return fmt.Errorf("can't read AK cert at %x: %w", certLocation, err)
	}

	akCert, err := x509.ParseCertificate(akCertBytes)
	if err != nil {
		return fmt.Errorf("can't parse AK cert: %w", err)
	}

	if akCert.PublicKeyAlgorithm.String() != "RSA" {
		return fmt.Errorf("Public key algorithm %s not supported", akCert.PublicKeyAlgorithm.String())
	}

	akPub := akCert.PublicKey.(*rsa.PublicKey)

	if debugLogging {
		log.Printf("AK cert:")
		log.Printf("\tSubject: %s", akCert.Subject)
		log.Printf("\tPubkey Alg: %s", akCert.PublicKeyAlgorithm.String())
		log.Printf("\t\tModulus: %x", akPub.N)
		log.Printf("\t\tExponent: %d", akPub.E)
		log.Printf("\tPubkey: %s", akCert.PublicKeyAlgorithm)
		log.Printf("\tIssuer: %s", akCert.Issuer)
		log.Printf("\tSignature: %x", akCert.Signature)
	}

	// Get the boot measurements
	bootMeasurements, err := os.ReadFile(bootMeasurementsLocation)
	if err != nil {
		return fmt.Errorf("couldn't read boot measurements: %w", err)
	}

	verityMeasurements, err := os.ReadFile(verityMeasurementsLocation)
	if err != nil {
		return fmt.Errorf("couldn't read verity measurements: %w", err)
	}

	if debugLogging {
		log.Printf("%d %d", len(bootMeasurements), len(verityMeasurements))
	}

	// Get a random nonce
	nonce := make([]byte, 8)
	rand.Read(nonce)

	// PCR_Read only supports reading 8 PCRs at a time
	// Select the first PCRs to quote
	pcrsel := tpm2.PCRSelection{
		Hash: tpm2.AlgSHA256,
		PCRs: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11},
	}
	quoteData, quoteSig, err := tpm2.QuoteRaw(rwc, tpmutil.Handle(akLocation), "", "", nonce, pcrsel, tpm2.AlgNull)
	if err != nil {
		return fmt.Errorf("couldn't quote PCRs: %w", err)
	}

	// PCR_Read only supports reading 8 PCRs at a time
	pcrsel = tpm2.PCRSelection{
		Hash: tpm2.AlgSHA256,
		PCRs: []int{0, 1, 2, 3, 4, 5, 6, 7},
	}
	pcrs, err := tpm2.ReadPCRs(rwc, pcrsel)
	if err != nil {
		return fmt.Errorf("couldn't read PCRs: %w", err)
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
		return fmt.Errorf("couldn't read PCRs: %w", err)
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

	if debugLogging {
		log.Printf("PCR Values:")
		for _, pcr := range pcrValues {
			log.Printf("\t%d: %x", pcr.Index, pcr.Value)
		}

		log.Printf("Quote Data: %x", quoteData)
		log.Printf("Quote Sig: %x", quoteSig)
	}

	// Generate the attestation to output
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
		return fmt.Errorf("couldn't serialize attestation: %w", err)
	}

	err = os.WriteFile(outputPath, json, 0666)
	if err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}
