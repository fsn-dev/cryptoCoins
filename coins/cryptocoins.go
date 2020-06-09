/*
 *  Copyright (C) 2018-2019  Fusion Foundation Ltd. All rights reserved.
 *  Copyright (C) 2018-2019  gaozhengxin@fusion.org
 *
 *  This library is free software; you can redistribute it and/or
 *  modify it under the Apache License, Version 2.0.
 *
 *  This library is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 *
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package coins

import (
	"math/big"
	"strings"

	"github.com/fsn-dev/cryptoCoins/coins/types"

	"github.com/fsn-dev/cryptoCoins/coins/atom"
	"github.com/fsn-dev/cryptoCoins/coins/bch"
	"github.com/fsn-dev/cryptoCoins/coins/bnb"
	"github.com/fsn-dev/cryptoCoins/coins/btc"
	"github.com/fsn-dev/cryptoCoins/coins/eos"
	"github.com/fsn-dev/cryptoCoins/coins/erc20"
	"github.com/fsn-dev/cryptoCoins/coins/eth"
	"github.com/fsn-dev/cryptoCoins/coins/evt"
	"github.com/fsn-dev/cryptoCoins/coins/fsn"
	"github.com/fsn-dev/cryptoCoins/coins/omni"
	"github.com/fsn-dev/cryptoCoins/coins/trx"
	"github.com/fsn-dev/cryptoCoins/coins/xrp"

	config "github.com/fsn-dev/cryptoCoins/coins/config"
)

var Coinmap map[string]string = make(map[string]string)

func Init() {
	for _, ct := range Cointypes {
		Coinmap[ct] = "1"
	}
	btc.BTCInit()
	eth.ETHInit()
	fsn.FSNInit()
	xrp.XRPInit()
	eos.EOSInit()
	erc20.ERC20Init()
	omni.OMNIInit()
	atom.ATOMInit()
	bch.BCHInit()
	evt.EVTInit()
	trx.TRXInit()
	erc20.RegisterTokenGetter(func(tokentype string) string {
		// TODO ¿¿¿ ???
		//return erc20.Tokens[tokentype]
		ret, ok := erc20.Tokens[tokentype]
		if ok {
			return ret
		} else {
			return ""
		}
	})
	omni.RegisterPropertyGetter(func(propertyname string) string {
		// TODO ¿¿¿ ???
		//return omni.Properties[propertyname]
		ret, ok := omni.Properties[propertyname]
		if ok {
			return ret
		} else {
			return ""
		}
	})
}

// only main net coins
//var Cointypes []string = []string{"ALL","BTC","ETH","XRP","EOS","USDT","ATOM","BCH","TRX","BNB","EVT1","ERC20BNB","ERC20GUSD","ERC20MKR","ERC20HT","ERC20RMBT","EVT1001","BEP2GZX_754"}
//var Cointypes []string = []string{"ALL","BTC","ETH","ATOM","BCH","TRX","BNB","ERC20BNB","ERC20GUSD","ERC20MKR","ERC20HT","ERC20RMBT","BEP2GZX_754"}  //tmp delete EOS XRP EVT1 EVT1001 USDT
//var Cointypes []string = []string{"ALL", "FSN", "ETH", "BTC", "ATOM", "BCH", "BNB", "EOS", "EVT1", "ERC20GUSD", "ERC20MKR", "ERC20HT", "ERC20BNB", "ERC20BNT", "ERC20RMBT", "TRX", "XRP", "ERC20USDT"}
var Cointypes []string = []string{"ALL", "FSN", "ETH", "BTC", "ATOM", "BCH", "BNB", "ERC20GUSD", "ERC20MKR", "ERC20HT", "ERC20BNB", "ERC20BNT", "ERC20RMBT", "TRX", "XRP", "ERC20USDT"}

//BEP2--->BEP2GZX_754

func IsCoinSupported(cointype string) bool {
	if cointype == "" {
		return false
	}

	cointype = strings.ToUpper(cointype)
	if Coinmap[cointype] == "1" {
		return true
	} else {
		//if strings.HasPrefix(cointype, "EVT") || strings.HasPrefix(cointype, "BEP2") || erc20.GetToken(cointype) != "" || omni.GetProperty(cointype) != "" {
		if strings.HasPrefix(cointype, "EVT") || strings.HasPrefix(cointype, "BEP2") {
			return true
		}

		if erc20.GetToken != nil && erc20.GetToken(cointype) != "" {
			return true
		}

		if omni.GetProperty != nil && omni.GetProperty(cointype) != "" {
			return true
		}
	}

	return false
}

type CryptocoinHandler interface {

	// 公钥to dcrm地址
	PublicKeyToAddress(pubKeyHex string) (address string, err error)

	// 构造未签名交易
	BuildUnsignedTransaction(fromAddress, fromPublicKey, toAddress string, amount *big.Int, jsonstring string, memo string) (transaction interface{}, digests []string, err error)

	// 签名函数 txhash 输出 rsv 测试用
	//SignTransaction(hash []string, privateKey interface{}) (rsv []string, err error)

	// 构造签名交易
	MakeSignedTransaction(rsv []string, transaction interface{}) (signedTransaction interface{}, err error)

	// 根据未签名交易的json数据构造签名交易
	MakeSignedTransactionByJson(rsv []string, txjson string) (signedTransaction interface{}, err error)

	// 提交交易
	SubmitTransaction(signedTransaction interface{}) (txhash string, err error)

	// 根据已签名交易的json数据得到签名交易并提交该交易
	SubmitTransactionByJson(txjson string) (txhash string, err error)

	// 根据交易hash查交易信息
	// fromAddress 交易发起方地址
	// txOutputs 交易输出切片, txOutputs[i].ToAddress 第i条交易接收方地址, txOutputs[i].Amount 第i条交易转账金额
	//GetTransactionInfo(txhash string) (fromAddress string, txOutputs []types.TxOutput, jsonstring string, confirmed bool, fee types.Value, err error)
	GetTransactionInfo(txhash string) (txinfo *types.TransactionInfo, err error)

	// 从区块中过滤特定交易
	FiltTransaction(blocknumber uint64, filter types.Filter) (txhashes []string, err error)

	// 账户查账户余额
	GetAddressBalance(address string, jsonstring string) (balance types.Balance, err error)

	// 默认交易费用
	GetDefaultFee() types.Value

	// 是coin还是token
	IsToken() bool
}

func NewCryptocoinHandler(coinType string) (txHandler CryptocoinHandler) {
	if !config.Loaded {
		config.Init()
	}
	coinTypeC := strings.ToUpper(coinType)
	switch coinTypeC {
	case "ATOM":
		return atom.NewAtomHandler()
	case "BTC":
		return btc.NewBTCHandler()
	case "ETH":
		return eth.NewETHHandler()
	case "FSN":
		return fsn.NewFSNHandler()
	case "XRP":
		return xrp.NewXRPHandler()
	case "EOS":
		return eos.NewEOSHandler()
	case "TRX":
		return trx.NewTRXHandler()
	case "BCH":
		return bch.NewBCHHandler()
	case "BNB":
		return bnb.NewBNBHandler()
	case "USDT":
		return omni.NewOMNIPropertyHandler("OMNIOmni") //testnet3测试网中的omni token, 可以代替USDT测试
	default:
		if IsErc20(coinTypeC) {
			return erc20.NewERC20TokenHandler(coinTypeC)
		}
		if IsOmni(coinTypeC) {
			return omni.NewOMNIPropertyHandler(coinType)
		}
		if IsEVT(coinTypeC) {
			return evt.NewEvtHandler(coinTypeC)
		}
		if IsBEP2(coinTypeC) {
			return bnb.NewBEP2Handler(coinTypeC)
		}
		return nil
	}
}

func IsEVT(tokentype string) bool {
	return strings.HasPrefix(tokentype, "EVT")
}

func IsErc20(tokentype string) bool {
	return strings.HasPrefix(tokentype, "ERC20")
}

func IsOmni(propertyname string) bool {
	return strings.HasPrefix(propertyname, "OMNI")
}

func IsBEP2(tokentype string) bool {
	return strings.HasPrefix(tokentype, "BEP2")
}

func GetMainNetCoin(cointype string) string {
	if IsEVT(cointype) {
		return "EVT1"
	}
	if IsErc20(cointype) {
		return "ETH"
	}
	if IsOmni(cointype) {
		return "BTC"
	}
	if IsBEP2(cointype) {
		return "BNB"
	}
	return cointype
}
