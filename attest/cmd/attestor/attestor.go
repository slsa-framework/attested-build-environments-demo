package main

import (
	"fmt"
	"os"

	"github.com/chkimes/image-attestation/internal"
	"github.com/in-toto/scai-demos/scai-gen/pkg/fileio"
	"github.com/in-toto/scai-demos/scai-gen/pkg/generators"

	ita "github.com/in-toto/attestation/go/v1"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "attestor",
	Args:  cobra.ExactArgs(1),
	Short: "Generates a JSON-encoded VM image attestation (in-toto format)",
	RunE:  genAttestation,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	outFile      string
	targetFile   string
	evidenceFile string
	prettyPrint  bool
)

func init() {
	rootCmd.Flags().StringVarP(
		&outFile,
		"out-file",
		"o",
		"",
		"Filename to write out the JSON-encoded object",
	)
	rootCmd.MarkFlagRequired("out-file") //nolint:errcheck

	rootCmd.Flags().StringVarP(
		&targetFile,
		"target",
		"t",
		"",
		"The filename of the JSON-encoded target attestation file",
	)

	rootCmd.Flags().StringVarP(
		&evidenceFile,
		"evidence",
		"e",
		"",
		"The filename of the JSON-encoded evidence file",
	)

	rootCmd.Flags().BoolVar(
		&prettyPrint,
		"pretty-print",
		false,
		"Flag to JSON pretty-print the generated Report",
	)
}

func genAttestation(_ *cobra.Command, args []string) error {
	// want to make sure the AttributeAssertion is a JSON file
	if !fileio.HasJSONExt(outFile) {
		return fmt.Errorf("expected a .json extension for the generated SCAI AttributeAssertion file %s", outFile)
	}

	// read in the subject file
	subjectFile := args[0]
	subject, err := generators.NewRdForFile(subjectFile, "vmID", "", "sha256", true, "application/x-pem-file", "examples/vm.id", nil)
	if err != nil {
		return fmt.Errorf("failed to generate RD for subject %s: %w", subjectFile, err)
	}

	target, err := generators.NewRdForFile(targetFile, "chkimes-image-attestation-attestation-675331.sigstore.json", "", "sha256", false, "application/vnd.dev.sigstore.bundle+json;version=0.2", "https://github.com/chkimes/image-attestation/attestations/675331/download", nil)

	st, err := internal.NewRefValueStatement([]*ita.ResourceDescriptor{subject}, "build_image", target, evidenceFile, false)
	if err != nil {
		return fmt.Errorf("failed to generate reference value: %w", err)
	}

	return fileio.WritePbToFile(st, outFile, prettyPrint)
}

func main() {
	Execute()
}
