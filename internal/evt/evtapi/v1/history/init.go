package history

import (
	"fmt"

	"github.com/fsn-dev/cryptoCoins/internal/evt/evtapi/client"
	"github.com/fsn-dev/cryptoCoins/internal/evt/evtconfig"
)

type Instance struct {
	Client *client.Instance
	Config *evtconfig.Instance
}

func New(config *evtconfig.Instance, client *client.Instance) *Instance {
	return &Instance{
		Client: client,
		Config: config,
	}
}

func (it *Instance) Path(method string) string {
	return fmt.Sprintf("history/%v", method)
}
