package cmd

import (
	"encoding/pem"
	"fmt"
	"os"

	"github.com/in-toto/scai-demos/scai-gen/pkg/fileio"

	sigbundle "github.com/sigstore/protobuf-specs/gen/pb-go/bundle/v1"
	"github.com/spf13/cobra"
)

var parseCmd = &cobra.Command{
	Use:   "parse-sigstore",
	Short: "Parses a JSON-encoded Sigstore bundle",
}

var pubkeyCmd = &cobra.Command{
	Use:   "pubkey",
	Args:  cobra.ExactArgs(1),
	Short: "Outputs the PEM-formatted public key contained in a Sigstore bundle body",
	RunE:  getPubKey,
}

func init() {
	pubkeyCmd.Flags().StringVarP(
		&outFile,
		"out-file",
		"o",
		"",
		"Filename to write out the JSON-encoded object",
	)

	parseCmd.AddCommand(pubkeyCmd)
}

func getPubKey(_ *cobra.Command, args []string) error {
	// read in the sigstore bundle file
	bundleFile := args[0]
	bundle := &sigbundle.Bundle{}

	err := fileio.ReadPbFromFile(bundleFile, bundle)
	if err != nil {
		return fmt.Errorf("failed to read Sigstore bundle file %s: %w", bundleFile, err)
	}

	certChain := bundle.GetVerificationMaterial().GetX509CertificateChain()
	if certChain == nil {
		return fmt.Errorf("failed to retrieve x509 certificatefrom Sigstore bundle %s", bundleFile)
	}

	certs := certChain.GetCertificates()
	if certs == nil || len(certs[0].GetRawBytes()) == 0 {
		return fmt.Errorf("failed to retrieve x509 leaf certificate from Sigstore bundle %s", bundleFile)
	}

	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certs[0].GetRawBytes(),
	}

	encoded := pem.EncodeToMemory(block)
	if encoded == nil {
		return fmt.Errorf("failed to PEM-encode x509 certificate for Sigstore bundle %s", bundleFile)
	}

	if len(outFile) > 0 {
		err = os.WriteFile(outFile, encoded, 0644)
	} else {
		fmt.Printf("Parsed: \n\n%s", string(encoded))
		err = nil
	}
	return err
}
