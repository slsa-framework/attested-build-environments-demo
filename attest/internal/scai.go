package internal

import (
	"fmt"

	"github.com/in-toto/scai-demos/scai-gen/pkg/generators"

	scai "github.com/in-toto/attestation/go/predicates/scai/v0"
	ita "github.com/in-toto/attestation/go/v1"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

func NewRefValueStatement(subjects []*ita.ResourceDescriptor, attribute string, target *ita.ResourceDescriptor, evidencePath string, includeEvidence bool) (*ita.Statement, error) {
	// generate the resource descriptor for the reference value evidence
	ev, err := generators.NewRdForFile(evidencePath, "", "", "sha256", includeEvidence, "application/json", "someLocation", nil)
	if err != nil {
		return nil, fmt.Errorf("error generating resource descriptor for evidence: %w", err)
	}

	// generate the SCAI assertion for the reference value
	scaiAA, err := generators.NewSCAIAssertion(attribute, target, nil, ev)
	if err != nil {
		return nil, fmt.Errorf("error generating SCAI assertion: %w", err)
	}

	// generate the in-toto predicate (SCAI report)
	scaiReport := &scai.AttributeReport{
		Attributes: []*scai.AttributeAssertion{scaiAA},
	}

	// plug the SCAI report into an in-toto Statement

	// sigh, this is the easiest way to get to a struct
	reportJSON, err := protojson.Marshal(scaiReport)
	if err != nil {
		return nil, fmt.Errorf("error marshalling SCAI report: %w", err)
	}

	reportStruct := &structpb.Struct{}
	err = protojson.Unmarshal(reportJSON, reportStruct)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling SCAI report: %w", err)
	}

	return generators.NewStatement(subjects, "https://in-toto.io/attestation/scai/attribute-report/v0.2", reportStruct)
}
