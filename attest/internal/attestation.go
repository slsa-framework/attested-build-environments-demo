package internal

type Attestation struct {
	AkCert         []byte     `json:"akCert"` // DER
	BootEventLog   []byte     `json:"bootEventLog"`
	VerityEventLog []byte     `json:"verityEventLog"`
	QuoteData      []byte     `json:"quoteData"`      // TPMS_ATTEST
	QuoteSignature []byte     `json:"quoteSignature"` // TPMT_SIGNATURE
	PCRs           []PCRValue `json:"pcrs"`
}

type PCRValue struct {
	Index int    `json:"index"`
	Value []byte `json:"value"`
}

type ExpectedPCRs struct {
	PCRs []PCRValue `json:"pcrs"`
}
