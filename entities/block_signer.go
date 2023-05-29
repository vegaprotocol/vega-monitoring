package entities

import (
	"time"

	vega_entities "code.vegaprotocol.io/vega/datanode/entities"
)

type BlockSigner struct {
	VegaTime time.Time                         `db:"vega_time"`
	Role     BlockSignerRole                   `db:"role"`
	TmPubKey vega_entities.TendermintPublicKey `db:"tendermint_pub_key"`
}
