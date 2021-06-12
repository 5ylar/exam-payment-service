package omiseprovider

import (
	"github.com/omise/omise-go"
	"github.com/omise/omise-go/operations"
)

type provider struct {
	oc *omise.Client
}

func New(oc *omise.Client) *provider {
	return &provider{
		oc,
	}
}

func (p *provider) CreateSource(createSource operations.CreateSource) (omise.Source, error) {
	source := &omise.Source{}

	if err := p.oc.Do(source, &createSource); err != nil {
		return *source, err
	}

	return *source, nil
}

func (p *provider) CreateCharge(createCharge operations.CreateCharge) (omise.Charge, error) {
	charge := &omise.Charge{}

	if err := p.oc.Do(charge, &createCharge); err != nil {
		return *charge, err
	}

	return *charge, nil
}
