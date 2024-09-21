package internal

import (
	"fmt"

	"github.com/in-toto/scai-demos/scai-gen/pkg/generators"

	scai "github.com/in-toto/attestation/go/predicates/scai/v0"
	ita "github.com/in-toto/attestation/go/v1"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

func NewRefValueSCAIAssertion(attribute string, targetPath string, includeTargetContent bool) (*scai.AttributeAssertion, error) {
	// generate the resource descriptor for the reference value target
	target, err := generators.NewRdForFile(targetPath, "", "", "sha256", includeTargetContent, "", "", nil)
	if err != nil {
		return nil, fmt.Errorf("error generating resource descriptor: %w", err)
	}

	// generate the SCAI assertion for the reference value
	scaiAA, err := generators.NewSCAIAssertion(attribute, target, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("error generating SCAI assertion: %w", err)
	}

	return scaiAA, nil
}

func NewSCAIStatement(subject []*ita.ResourceDescriptor, attributeAssertions []*scai.AttributeAssertion, producer *ita.ResourceDescriptor) (*ita.Statement, error) {
	// create the in-toto predicate (SCAI report)
	scaiReport := &scai.AttributeReport{
		Attributes: attributeAssertions,
		Producer:   producer,
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

	return generators.NewStatement(subject, "https://in-toto.io/attestation/scai/attribute-report/v0.2", reportStruct)
}
