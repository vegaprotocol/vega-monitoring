package entities

import (
	"time"

	vega_entities "code.vegaprotocol.io/vega/datanode/entities"
)

type BlockSigner struct {
	VegaTime  time.Time                          `db:"vega_time"`
	Height    int64                              `db:"height"`
	Role      BlockSignerRole                    `db:"role"`
	TmAddress string                             `db:"tendermint_address"`
	TmPubKey  *vega_entities.TendermintPublicKey `db:"tendermint_pub_key"`
}
