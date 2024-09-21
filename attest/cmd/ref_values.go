package cmd

import (
	"fmt"

	"github.com/chkimes/image-attestation/internal"
	//"github.com/in-toto/scai-demos/scai-gen/pkg/fileio"
	"github.com/in-toto/scai-demos/scai-gen/pkg/generators"

	scai "github.com/in-toto/attestation/go/predicates/scai/v0"
	ita "github.com/in-toto/attestation/go/v1"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

var refValuesCmd = &cobra.Command{
	Use:   "ref-values",
	Short: "Generates the reference values of build environment components and outputs a DSSE-signed in-toto SCAI attestation",
	RunE:  genRefValues,
}

var (
	buildImgFile     string
	kernelFile       string
	initramfsFile    string
	verityFile       string
	vmmPcrsFile      string
	previewRefValues bool
	prettyPrint      bool
)

func init() {
	refValuesCmd.Flags().StringVarP(
		&outFile,
		"out-file",
		"o",
		"ref-values.jsonl",
		"Filename to write out the JSON-encoded DSSE object",
	)

	refValuesCmd.Flags().StringVarP(
		&buildImgFile,
		"build-image-file",
		"b",
		"",
		"The name of the bundled build image",
	)

	refValuesCmd.Flags().StringVarP(
		&kernelFile,
		"kernel-file",
		"k",
		"",
		"The name of the kernel image",
	)

	refValuesCmd.Flags().StringVarP(
		&initramfsFile,
		"initramfs-file",
		"i",
		"",
		"The name of the initramfs file",
	)

	refValuesCmd.Flags().StringVarP(
		&verityFile,
		"verity-file",
		"v",
		"",
		"The name of the verity tree file",
	)

	refValuesCmd.Flags().StringVarP(
		&vmmPcrsFile,
		"vmm-pcrs-file",
		"p",
		"",
		"The name of the expected VMM-set PCR values file",
	)

	refValuesCmd.Flags().BoolVar(
		&previewRefValues,
		"preview-ref-values",
		true,
		"Flag to display the unsigned in-toto SCAI attestation",
	)

	refValuesCmd.Flags().BoolVar(
		&prettyPrint,
		"pretty-print",
		false,
		"Flag to JSON pretty-print the generated Report",
	)
}

func genRefValues(_ *cobra.Command, args []string) error {

	// Generate SCAI attribute assertions for each measured build environment component

	kernelRef, err := internal.NewRefValueSCAIAssertion("REF_VALUE:kernel", kernelFile, false)
	if err != nil {
		return fmt.Errorf("failed to generate SCAI assertion for the kernel %s: %w", kernelFile, err)
	}

	initramfsRef, err := internal.NewRefValueSCAIAssertion("REF_VALUE:initramfs", initramfsFile, false)
	if err != nil {
		return fmt.Errorf("failed to generate SCAI assertion for the initramfs %s: %w", initramfsFile, err)
	}

	verityRef, err := internal.NewRefValueSCAIAssertion("REF_VALUE:verity-hash", verityFile, false)
	if err != nil {
		return fmt.Errorf("failed to generate SCAI assertion for the verity hash %s: %w", verityFile, err)
	}

	vmmPcrsRef, err := internal.NewRefValueSCAIAssertion("REF_VALUE:vmm-pcrs", vmmPcrsFile, true)
	if err != nil {
		return fmt.Errorf("failed to generate SCAI assertion for the VMM-set PCRs %s: %w", vmmPcrsFile, err)
	}

	// The build image file is the subject of the in-toto attestation
	subject, err := generators.NewRdForFile(buildImgFile, "", "", "sha256", false, "", "", nil)
	if err != nil {
		return fmt.Errorf("failed to generate RD for the build image %s: %w", buildImgFile, err)
	}

	statement, err := internal.NewSCAIStatement([]*ita.ResourceDescriptor{subject}, []*scai.AttributeAssertion{kernelRef, initramfsRef, verityRef, vmmPcrsRef}, nil)
	if err != nil {
		return fmt.Errorf("failed to generate in-toto Statement for SCAI predicate: %w", err)
	}

	if previewRefValues {
		fmt.Printf("%s\n", protojson.Format(statement))
	}

	// TODO: Implement DSSE signing
	fmt.Printf("CAUTION: DSSE signing not implemented yet. NOT READY FOR PRODUCTION!!\n")

	return nil

	//return fileio.WritePbToFile(st, outFile, prettyPrint)
}
