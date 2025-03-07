module github.com/chkimes/image-attestation

go 1.23.0

toolchain go1.24.1

require (
	github.com/google/go-tpm v0.9.3
	github.com/in-toto/attestation v1.1.1
	github.com/in-toto/scai-demos v0.3.0
	github.com/sigstore/protobuf-specs v0.4.0
	github.com/spf13/cobra v1.9.1
	golang.org/x/exp v0.0.0-20250305212735-054e65f0b394
	google.golang.org/protobuf v1.36.5
)

require (
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/google/cel-go v0.20.1 // indirect
	github.com/in-toto/attestation-verifier v0.0.0-20231007025621-3193280f5194 // indirect
	github.com/in-toto/in-toto-golang v0.9.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	github.com/secure-systems-lab/go-securesystemslib v0.8.0 // indirect
	github.com/shibumi/go-pathspec v1.3.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/stoewer/go-strcase v1.2.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240311173647-c811ad7063a7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/chkimes/image-attestation => /home/mmelara/build-env-attestation/attest
