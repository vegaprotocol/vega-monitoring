package entities

import (
	"fmt"
)

type BlockSignerRole string

const (
	BlockSignerRoleProposer BlockSignerRole = "ROLE_PROPOSER"
	BlockSignerRoleSigner   BlockSignerRole = "ROLE_SIGNER"
)

func (n BlockSignerRole) IsValid() error {
	switch n {
	case BlockSignerRoleProposer, BlockSignerRoleSigner:
		return nil
	}
	return fmt.Errorf("Invalid Block Signer Role %s", n)
}
