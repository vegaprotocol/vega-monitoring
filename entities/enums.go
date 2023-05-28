package entities

import (
	"encoding/base64"
	"fmt"

	"github.com/jackc/pgtype"
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

type TendermintAddress string

func (pk *TendermintAddress) Bytes() ([]byte, error) {
	strPK := pk.String()

	bytes, err := base64.StdEncoding.DecodeString(strPK)
	if err != nil {
		return nil, fmt.Errorf("decoding '%v': %w", pk.String(), err)
	}
	return bytes, nil
}

func (pk *TendermintAddress) String() string {
	return string(*pk)
}

func (pk TendermintAddress) EncodeBinary(ci *pgtype.ConnInfo, buf []byte) ([]byte, error) {
	bytes, err := pk.Bytes()
	if err != nil {
		return buf, err
	}
	return append(buf, bytes...), nil
}
