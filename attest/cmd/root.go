package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "image-attestation",
	Short: "A CLI tool for the SLSA Attested Build Environments track",
}

var (
	outFile string
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var debugLogging bool

func init() {
	rootCmd.AddCommand(quoteCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(refValuesCmd)
	//rootCmd.AddCommand(parseCmd)
}

func main() {
	Execute()
}
