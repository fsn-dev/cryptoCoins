package evt

import (
	"github.com/fsn-dev/cryptoCoins/internal/evt/evtapi"
	"github.com/fsn-dev/cryptoCoins/internal/evt/evtconfig"
	"github.com/sirupsen/logrus"
)

type Instance struct {
	Api *evtapi.Instance
	Log *logrus.Logger
}

func (it *Instance) SilenceLog() {
	it.Log.SetLevel(logrus.ErrorLevel)
}

func New(config *evtconfig.Instance) *Instance {
	logger := logrus.New()
	return &Instance{
		Api: evtapi.New(config, logger),
		Log: logger,
	}
}
