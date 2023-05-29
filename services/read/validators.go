package read

import "github.com/vegaprotocol/data-metrics-store/clients/comet"

func (s *ReadService) GetValidatorForAddressAtBlock(address string, block int64) (*comet.ValidatorData, error) {
	return s.cometClient.GetValidatorForAddressAtBlock(address, block)
}
