module github.com/chkimes/image-attestation

go 1.21

toolchain go1.21.5

require (
	github.com/google/go-tpm v0.9.0
	github.com/in-toto/attestation v1.0.1
	github.com/in-toto/scai-demos v0.3.0
	github.com/spf13/cobra v1.8.0
	golang.org/x/exp v0.0.0-20240205201215-2c58cdc269a3
	google.golang.org/protobuf v1.33.0
)

require (
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/google/cel-go v0.20.1 // indirect
	github.com/in-toto/attestation-verifier v0.0.0-20231007025621-3193280f5194 // indirect
	github.com/in-toto/in-toto-golang v0.9.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/secure-systems-lab/go-securesystemslib v0.8.0 // indirect
	github.com/shibumi/go-pathspec v1.3.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stoewer/go-strcase v1.2.0 // indirect
	golang.org/x/crypto v0.20.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240102182953-50ed04b92917 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240116215550-a9fa1716bcac // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/chkimes/image-attestation => /home/mmelara/build-env-attestation/attest
