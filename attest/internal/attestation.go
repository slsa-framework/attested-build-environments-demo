package internal

type Attestation struct {
	AkCert         []byte `json:"akCert"` // DER
	BootEventLog   []byte `json:"bootEventLog"`
	VerityEventLog []byte `json:"verityEventLog"`
	QuoteData      []byte `json:"quoteData"`      // TPMS_ATTEST
	QuoteSignature []byte `json:"quoteSignature"` //
}
