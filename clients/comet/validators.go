package comet

import (
	"fmt"
	"strconv"

	vega_entities "code.vegaprotocol.io/vega/datanode/entities"
)

//
//
//

func (c *CometClient) GetValidatorForAddressAtBlock(address string, block int64) (*ValidatorData, error) {
	if val, ok := c.validatorByAddress[address]; ok {
		return &val, nil
	}
	if err := c.pullMoreValidatorData(block); err != nil {
		return nil, fmt.Errorf("failed to get validator data for %s address, failed to pull more validator data, %w", address, err)
	}
	if val, ok := c.validatorByAddress[address]; ok {
		return &val, nil
	}
	return nil, fmt.Errorf("failed to get validator data for %s address", address)
}

func (c *CometClient) pullMoreValidatorData(block int64) error {
	response, err := c.requestValidators(block)
	if err != nil {
		return err
	}
	for _, val := range response.Result.Validators {
		if _, ok := c.validatorByAddress[val.Address]; !ok {
			c.validatorByAddress[val.Address] = ValidatorData{
				Address:  val.Address,
				TmPubKey: vega_entities.TendermintPublicKey(val.PubKey.Value),
			}
		}
	}
	return nil
}

//
//
//

type CometValidators struct {
	Address          string
	TmPubKey         vega_entities.TendermintPublicKey
	VotingPower      int64
	ProposerPriority int64
	Height           int64
}

func (c *CometClient) GetValidators() ([]CometValidators, error) {
	return c.GetValidatorsForBlock(0)
}

func (c *CometClient) GetValidatorsForBlock(block int64) ([]CometValidators, error) {
	response, err := c.requestValidators(block)
	if err != nil {
		return nil, err
	}
	validatorList, err := newCometValidators(response)
	if err != nil {
		return nil, err
	}
	return validatorList, nil
}

func newCometValidators(response validatorsResponse) (cometValidators []CometValidators, err error) {
	height, err := strconv.ParseInt(response.Result.BlockHeight, 10, 64)
	if err != nil {
		err = fmt.Errorf("failed to parse Height '%s' to int, from: %+v.", response.Result.BlockHeight, response)
		return
	}

	for _, val := range response.Result.Validators {
		var votingPower, proposerPriority int64
		// Voting Power
		votingPower, err = strconv.ParseInt(val.VotingPower, 10, 64)
		if err != nil {
			err = fmt.Errorf("failed to parse VotingPower to int, from: %+v, for block %d, %w", val, height, err)
			return
		}
		// Proposer Priority
		proposerPriority, err = strconv.ParseInt(val.ProposerPriority, 10, 64)
		if err != nil {
			err = fmt.Errorf("failed to parse ProposerPriority to int, from: %+v, for block %d, %w", val, height, err)
			return
		}

		cometValidators = append(cometValidators, CometValidators{
			Address:          val.Address,
			TmPubKey:         vega_entities.TendermintPublicKey(val.PubKey.Value),
			VotingPower:      votingPower,
			ProposerPriority: proposerPriority,
			Height:           height,
		})
	}
	return
}
