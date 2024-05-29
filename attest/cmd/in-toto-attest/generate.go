package main

import (
	"fmt"

	"github.com/chkimes/image-attestation/internal"
	"github.com/in-toto/scai-demos/scai-gen/pkg/fileio"
	"github.com/in-toto/scai-demos/scai-gen/pkg/generators"

	ita "github.com/in-toto/attestation/go/v1"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Args:  cobra.ExactArgs(1),
	Short: "Generates a JSON-encoded in-toto Statement (unsigned) for a predicate about a given subject artifact",
	RunE:  genStatement,
}

var (
	targetFile   string
	evidenceFile string
	prettyPrint  bool
)

func init() {
	generateCmd.Flags().StringVarP(
		&outFile,
		"out-file",
		"o",
		"",
		"Filename to write out the JSON-encoded object",
	)
	generateCmd.MarkFlagRequired("out-file") //nolint:errcheck

	generateCmd.Flags().StringVarP(
		&targetFile,
		"target",
		"t",
		"",
		"The filename of the JSON-encoded target attestation file",
	)

	generateCmd.Flags().StringVarP(
		&evidenceFile,
		"evidence",
		"e",
		"",
		"The filename of the JSON-encoded evidence file",
	)

	generateCmd.Flags().BoolVar(
		&prettyPrint,
		"pretty-print",
		false,
		"Flag to JSON pretty-print the generated Report",
	)
}

func genStatement(_ *cobra.Command, args []string) error {
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
