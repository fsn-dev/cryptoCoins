package btc

import (
	"encoding/json"
	"fmt"
	"math/big"
	"runtime/debug"
	"sort"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/fsn-dev/cryptoCoins/coins/config"
	rpcutils "github.com/fsn-dev/cryptoCoins/coins/rpcutils"
	"github.com/fsn-dev/cryptoCoins/tools/common"
)

func ListUnspent_electrs(addr string) (list []btcjson.ListUnspentResult, balance *big.Int, err error) {
	return listUnspent_electrs(addr)
}

func listUnspent_electrs(addr string) (list []btcjson.ListUnspentResult, balance *big.Int, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("Runtime error: %v\n%v", e, string(debug.Stack()))
			return
		}
	}()
	path := `address/` + addr + `/utxo`
	ret, err := rpcutils.HttpGet(config.ApiGateways.BitcoinGateway.ElectrsAddress, path, nil)
	if err != nil {
		return
	}
	var utxos []electrsUtxo
	err = json.Unmarshal(ret, &utxos)
	if err != nil {
		return
	}

	fmt.Printf("\n\n%v\n\n", string(ret))
	fmt.Printf("\n\n%+v\n\n", utxos)

	balance = big.NewInt(0)

	for _, utxo := range utxos {
		balance = balance.Add(balance, big.NewInt(int64(utxo.Value)))
		path = `tx/` + utxo.Txid
		txret, txerr := rpcutils.HttpGet(config.ApiGateways.BitcoinGateway.ElectrsAddress, path, nil)
		if txerr != nil {
			fmt.Printf("%v ======== listUnspent_electrs,get utxo script,err = %v ========\n", common.CurrentTime(), txerr)
			continue
		}
		var tx electrsTx
		txerr = json.Unmarshal(txret, &tx)
		if txerr != nil {
			fmt.Printf("%v ======== listUnspent_electrs,get utxo script,err = %v ========\n", common.CurrentTime(), txerr)
			continue
		}
		utxo.Script = tx.Vout[int(utxo.Vout)].Scriptpubkey
		res := btcjson.ListUnspentResult{
			TxID:         utxo.Txid,
			Vout:         uint32(utxo.Vout),
			ScriptPubKey: utxo.Script,
			Address:      addr,
			Amount:       utxo.Value / 1e8,
			Spendable:    true,
		}
		if utxo.Status.Confirmed {
			res.Confirmations = 6
		} else {
			res.Confirmations = 0
		}
		list = append(list, res)
	}
	sort.Sort(sortableLURSlice(list))
	return
}

func GetTransaction_electrs(hash string) (*electrsTx, error) {
	path := `tx/` + hash
	txret, txerr := rpcutils.HttpGet(config.ApiGateways.BitcoinGateway.ElectrsAddress, path, nil)
	if txerr != nil {
		fmt.Printf("%v ======== GetTransaction_electrs,get utxo script,err = %v ========\n", common.CurrentTime(), txerr)
		return nil, txerr
	}
	var tx electrsTx
	txerr = json.Unmarshal(txret, &tx)
	if txerr != nil {
		fmt.Printf("%v ======== GetTransaction_electrs,get utxo script,err = %v ========\n", common.CurrentTime(), txerr)
		return nil, txerr
	}
	return &tx, nil
}

type electrsTx struct {
	Txid string
	Vout []electrsTxOut
	Fee  float64
}

type electrsTxOut struct {
	Scriptpubkey string
}

type electrsUtxo struct {
	Txid   string `json:"txid"`
	Vout   uint32
	Script string
	Status utxoStatus
	Value  float64
}

type utxoStatus struct {
	Confirmed    bool
	Block_height float64
	Block_hash   string
	Block_time   float64
}
