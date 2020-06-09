/*
 *  Copyright (C) 2018-2019  Fusion Foundation Ltd. All rights reserved.
 *  Copyright (C) 2018-2019  caihaijun@fusion.org huangweijun@fusion.org
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

package main

import (
	"fmt"
	"math/big"
	"net"
	"os"

	"github.com/fsn-dev/cryptoCoins/coins"
	rpc "github.com/fsn-dev/cryptoCoins/tools/rpcservice"
	"gopkg.in/urfave/cli.v1"
	"os/signal"
	"strconv"
	"strings"
	//"github.com/fsn-dev/cryptoCoins/tools/common"
	"encoding/json"
	cryptocoinsconfig "github.com/fsn-dev/cryptoCoins/coins/config"
	"github.com/fsn-dev/cryptoCoins/coins/types"
)

func main() {

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func StartCoins(c *cli.Context) {
	cryptocoinsconfig.Init()
	coins.Init()
	RpcInit(rpcport)
	select {} // note for server, or for client
}

var (
	rpcport int
	app     = cli.NewApp()
)

func init() {
	app.Usage = "Crypto Coins Service"
	app.Version = "0.0"
	app.Action = StartCoins
	app.Flags = []cli.Flag{
		cli.IntFlag{Name: "rpcport", Value: 0, Usage: "listen port", Destination: &rpcport},
	}
}

//===================================================================

func listenSignal(exit chan int) {
	sig := make(chan os.Signal)
	signal.Notify(sig)

	fmt.Println("============call listenSignal=============")
	for {
		<-sig
		exit <- 1
	}
}

type Service struct{}

var (
	rpc_port int
	endpoint string = "0.0.0.0"
	server   *rpc.Server
	err      error
)

func RpcInit(port int) {
	rpc_port = port
	go startRpcServer()
}

// splitAndTrim splits input separated by a comma
// and trims excessive white space from the substrings.
func splitAndTrim(input string) []string {
	result := strings.Split(input, ",")
	for i, r := range result {
		result[i] = strings.TrimSpace(r)
	}
	return result
}

func startRpcServer() error {
	go func() error {
		server = rpc.NewServer()
		service := new(Service)
		if err := server.RegisterName("coins", service); err != nil {
			panic(err)
		}

		// All APIs registered, start the HTTP listener
		var (
			listener net.Listener
			err      error
		)

		endpoint = endpoint + ":" + strconv.Itoa(rpc_port)
		if listener, err = net.Listen("tcp", endpoint); err != nil {
			panic(err)
		}

		vhosts := make([]string, 0)
		cors := splitAndTrim("*")
		go rpc.NewHTTPServer(cors, vhosts, rpc.DefaultHTTPTimeouts, server).Serve(listener)
		rpcstring := "\n==================== RPC Service Start! url = " + fmt.Sprintf("http://%s", endpoint) + " =====================\n"
		fmt.Println(rpcstring)

		exit := make(chan int)
		<-exit

		server.Stop()

		return nil
	}()

	return nil
}

//=============================================================================

type DcrmAddrRes struct {
	PubKey   string
	DcrmAddr string
	Cointype string
}

type DcrmPubkeyRes struct {
	PubKey      string
	DcrmAddress map[string]string
}

func GetAddr(pubkey string, cointype string) (string, string, error) {
	var m interface{}
	if !strings.EqualFold(cointype, "ALL") {

		h := coins.NewCryptocoinHandler(cointype)
		if h == nil {
			return "", "cointype is not supported", fmt.Errorf("req addr fail.cointype is not supported.")
		}

		ctaddr, err := h.PublicKeyToAddress(pubkey)
		if err != nil {
			return "", "dcrm back-end internal error:get dcrm addr fail from pubkey:" + pubkey, fmt.Errorf("req addr fail.")
		}

		m = &DcrmAddrRes{PubKey: pubkey, DcrmAddr: ctaddr, Cointype: cointype}
		b, _ := json.Marshal(m)
		return string(b), "", nil
	}

	addrmp := make(map[string]string)
	for _, ct := range coins.Cointypes {
		if strings.EqualFold(ct, "ALL") {
			continue
		}

		h := coins.NewCryptocoinHandler(ct)
		if h == nil {
			continue
		}
		ctaddr, err := h.PublicKeyToAddress(pubkey)
		if err != nil {
			continue
		}

		addrmp[ct] = ctaddr
	}

	m = &DcrmPubkeyRes{PubKey: pubkey, DcrmAddress: addrmp}
	b, _ := json.Marshal(m)
	return string(b), "", nil
}

func (this *Service) GetAddr(pubkey string, cointype string) map[string]interface{} {
	fmt.Printf("=====================call rpc GetAddr, pubkey = %v, cointype = %v ===========================\n", pubkey, cointype)
	data := make(map[string]interface{})
	if pubkey == "" || cointype == "" {
		data["result"] = ""
		return map[string]interface{}{
			"Status": "Error",
			"Tip":    "param error",
			"Error":  "param error",
			"Data":   data,
		}
	}

	ret, tip, err := GetAddr(pubkey, cointype)
	if err != nil {
		data["result"] = ""
		return map[string]interface{}{
			"Status": "Error",
			"Tip":    tip,
			"Error":  err.Error(),
			"Data":   data,
		}
	}

	data["result"] = ret
	return map[string]interface{}{
		"Status": "Success",
		"Tip":    "",
		"Error":  "",
		"Data":   data,
	}
}

func GetTransactionInfo(txhash string, cointype string) (*types.TransactionInfo, string, error) {
	h := coins.NewCryptocoinHandler(cointype)
	if h == nil {
		return nil, "unsupported cointype", fmt.Errorf("unsupported cointype")
	}

	txinfo, err := h.GetTransactionInfo(txhash)
	return txinfo, "", err
}

func (this *Service) GetTransactionInfo(txhash string, cointype string) map[string]interface{} {
	fmt.Printf("=====================call rpc GetTransactionInfo, txhash = %v, cointype = %v ===========================\n", txhash, cointype)
	data := make(map[string]interface{})
	if txhash == "" || cointype == "" {
		data["result"] = ""
		return map[string]interface{}{
			"Status": "Error",
			"Tip":    "param error",
			"Error":  "param error",
			"Data":   data,
		}
	}

	ret, tip, err := GetTransactionInfo(txhash, cointype)
	if err != nil {
		data["result"] = ""
		return map[string]interface{}{
			"Status": "Error",
			"Tip":    tip,
			"Error":  err.Error(),
			"Data":   data,
		}
	}

	data["result"] = ret
	return map[string]interface{}{
		"Status": "Success",
		"Tip":    "",
		"Error":  "",
		"Data":   data,
	}
}

type UnsignTx struct {
	Tx     interface{}
	TxHash []string
}

func BuildUnsignedTransaction(fromaddr string, pubkey string, toaddr string, amount string, memo string, cointype string) (*UnsignTx, string, error) {
	h := coins.NewCryptocoinHandler(cointype)
	if h == nil {
		return nil, "unsupported cointype", fmt.Errorf("unsupported cointype")
	}

	value, ok := new(big.Int).SetString(amount, 10)
	if ok == false {
		return nil, "get value error", fmt.Errorf("get value error")
	}

	tx, txhash, err := h.BuildUnsignedTransaction(fromaddr, pubkey, toaddr, value, "", memo)
	//b, _ := json.Marshal(tx)
	ut := &UnsignTx{Tx: tx, TxHash: txhash}
	return ut, "", err
}

func (this *Service) BuildUnsignedTransaction(fromaddr string, pubkey string, toaddr string, amount string, memo string, cointype string) map[string]interface{} {
	fmt.Printf("=====================call rpc BuildUnsignedTransaction, fromaddr = %v, pubkey = %v, toaddr = %v, amount = %v, memo = %v, cointype = %v ===========================\n", fromaddr, pubkey, toaddr, amount, memo, cointype)
	data := make(map[string]interface{})
	if fromaddr == "" || pubkey == "" || toaddr == "" || amount == "" || cointype == "" {
		data["result"] = ""
		return map[string]interface{}{
			"Status": "Error",
			"Tip":    "param error",
			"Error":  "param error",
			"Data":   data,
		}
	}

	ret, tip, err := BuildUnsignedTransaction(fromaddr, pubkey, toaddr, amount, memo, cointype)
	if err != nil {
		data["result"] = ""
		return map[string]interface{}{
			"Status": "Error",
			"Tip":    tip,
			"Error":  err.Error(),
			"Data":   data,
		}
	}

	data["result"] = ret
	return map[string]interface{}{
		"Status": "Success",
		"Tip":    "",
		"Error":  "",
		"Data":   data,
	}
}

func MakeSignedTransaction(txjson string, rsv string, cointype string) (string, string, error) {
	h := coins.NewCryptocoinHandler(cointype)
	if h == nil {
		return "", "unsupported cointype", fmt.Errorf("unsupported cointype")
	}

	/*var tx types.Transaction
	    err := json.Unmarshal([]byte(txjson), &tx)
	    if err != nil {
		fmt.Printf("==================MakeSignedTransaction,unmarshal txjson,err = %v ====================\n",err)
		return "","",err
	    }
	    fmt.Printf("==================MakeSignedTransaction,end unmarshal txjson, tx = %v ====================\n",tx)
	    tx2,ok := tx.(map[string]interface{})
	    if ok == false {
		return "","",fmt.Errorf("tx json data error")
	    }
	    s := fmt.Sprintf("%v",tx2["to"])
	    to,ok := new(big.Int).SetString(s,0)
	    if ok == false {
		return "","",fmt.Errorf("tx json data error")
	    }
	    toaddr := common.BigToAddress(to)
	    s = fmt.Sprintf("%v",tx2["input"])
	    input,ok := new(big.Int).SetString(s,0)
	    if ok == false {
		return "","",fmt.Errorf("tx json data error")
	    }
	    data := input.Bytes()
	    s = fmt.Sprintf("%v",tx2["nonce"])
	    nonce,ok := new(big.Int).SetString(s,0)
	    if ok == false {
		return "","",fmt.Errorf("tx json data error")
	    }
	    n := nonce.Uint64()
	    s = fmt.Sprintf("%v",tx2["gasPrice"])
	    gasprice,ok := new(big.Int).SetString(s,0)
	    if ok == false {
		return "","",fmt.Errorf("tx json data error")
	    }
	    s = fmt.Sprintf("%v",tx2["gas"])
	    gas,ok := new(big.Int).SetString(s,0)
	    if ok == false {
		return "","",fmt.Errorf("tx json data error")
	    }
	    gaslimit := gas.Uint64()
	    s = fmt.Sprintf("%v",tx2["value"])
	    value,ok := new(big.Int).SetString(s,0)
	    if ok == false {
		return "","",fmt.Errorf("tx json data error")
	    }
	    txtmp := types.NewTransaction(n,toaddr,value,gaslimit,gasprice,data)*/

	rsvs := make([]string, 0)
	rsvs = append(rsvs, rsv)
	//fmt.Printf("==================MakeSignedTransaction,start make signed tx,nonce = %v,gasprice = %v,gas = %v,to = %v,value = %v,input = %v,hash = %v ====================\n",n,gasprice,gaslimit,toaddr.Hex(),value,string(data),txtmp.Hash().Hex())
	signtx, err := h.MakeSignedTransactionByJson(rsvs, txjson)
	if err != nil {
		fmt.Printf("==================MakeSignedTransaction,make signed tx fail,err = %v ====================\n", err)
		return "", "", err
	}
	b, err := json.Marshal(signtx)
	if err != nil {
		fmt.Printf("==================MakeSignedTransaction,marshal signed tx fail,err = %v ====================\n", err)
		return "", "", err
	}
	return string(b), "", err
}

func (this *Service) MakeSignedTransaction(tx string, rsv string, cointype string) map[string]interface{} {
	fmt.Printf("=====================call rpc MakeSignedTransaction, tx = %v, rsv = %v, cointype = %v ===========================\n", tx, rsv, cointype)
	data := make(map[string]interface{})
	if tx == "" || rsv == "" || cointype == "" {
		data["result"] = ""
		return map[string]interface{}{
			"Status": "Error",
			"Tip":    "param error",
			"Error":  "param error",
			"Data":   data,
		}
	}

	rsvs := []rune(rsv)
	if string(rsvs[0:2]) == "0x" {
		rsv = string(rsvs[2:])
	}

	ret, tip, err := MakeSignedTransaction(tx, rsv, cointype)
	fmt.Printf("=====================finish call rpc MakeSignedTransaction, ret = %v, err = %v, ===========================\n", ret, err)
	if err != nil {
		data["result"] = ""
		return map[string]interface{}{
			"Status": "Error",
			"Tip":    tip,
			"Error":  err.Error(),
			"Data":   data,
		}
	}

	data["result"] = ret
	return map[string]interface{}{
		"Status": "Success",
		"Tip":    "",
		"Error":  "",
		"Data":   data,
	}
}

func SubmitTransaction(signtx string, cointype string) (string, string, error) {
	h := coins.NewCryptocoinHandler(cointype)
	if h == nil {
		return "", "unsupported cointype", fmt.Errorf("unsupported cointype")
	}

	/*var tx interface{}
	    err := json.Unmarshal([]byte(signtx), &tx)
	    if err != nil {
		fmt.Printf("==================SubmitTransaction,unmarshal txjson,err = %v ====================\n",err)
		return "","",err
	    }
	    tx2,ok := tx.(map[string]interface{})
	    if ok == false {
		return "","",fmt.Errorf("tx json data error")
	    }
	    s := fmt.Sprintf("%v",tx2["to"])
	    to,ok := new(big.Int).SetString(s,0)
	    if ok == false {
		return "","",fmt.Errorf("tx json data error")
	    }
	    toaddr := common.BigToAddress(to)
	    s = fmt.Sprintf("%v",tx2["input"])
	    input,ok := new(big.Int).SetString(s,0)
	    if ok == false {
		return "","",fmt.Errorf("tx json data error")
	    }
	    data := input.Bytes()
	    s = fmt.Sprintf("%v",tx2["nonce"])
	    nonce,ok := new(big.Int).SetString(s,0)
	    if ok == false {
		return "","",fmt.Errorf("tx json data error")
	    }
	    n := nonce.Uint64()
	    s = fmt.Sprintf("%v",tx2["gasPrice"])
	    gasprice,ok := new(big.Int).SetString(s,0)
	    if ok == false {
		return "","",fmt.Errorf("tx json data error")
	    }
	    s = fmt.Sprintf("%v",tx2["gas"])
	    gas,ok := new(big.Int).SetString(s,0)
	    if ok == false {
		return "","",fmt.Errorf("tx json data error")
	    }
	    gaslimit := gas.Uint64()
	    s = fmt.Sprintf("%v",tx2["value"])
	    value,ok := new(big.Int).SetString(s,0)
	    if ok == false {
		return "","",fmt.Errorf("tx json data error")
	    }
	    txtmp := types.NewTransaction(n,toaddr,value,gaslimit,gasprice,data)
	*/

	txhash, err := h.SubmitTransactionByJson(signtx)
	return txhash, "", err
}

func (this *Service) SubmitTransaction(signtx string, cointype string) map[string]interface{} {
	fmt.Printf("=====================call rpc ,SubmitTransaction, signtx = %v, cointype = %v ===========================\n", signtx, cointype)
	data := make(map[string]interface{})
	if signtx == "" || cointype == "" {
		data["result"] = ""
		return map[string]interface{}{
			"Status": "Error",
			"Tip":    "param error",
			"Error":  "param error",
			"Data":   data,
		}
	}

	ret, tip, err := SubmitTransaction(signtx, cointype)
	if err != nil {
		data["result"] = ""
		return map[string]interface{}{
			"Status": "Error",
			"Tip":    tip,
			"Error":  err.Error(),
			"Data":   data,
		}
	}

	data["result"] = ret
	return map[string]interface{}{
		"Status": "Success",
		"Tip":    "",
		"Error":  "",
		"Data":   data,
	}
}

func (this *Service) GetDefaultFee(cointype string) map[string]interface{} {
	fmt.Printf("=====================call rpc GetDefaultFee, cointype = %v ===========================\n", cointype)
	data := make(map[string]interface{})
	if cointype == "" {
		data["result"] = ""
		return map[string]interface{}{
			"Status": "Error",
			"Tip":    "param error",
			"Error":  "param error",
			"Data":   data,
		}
	}

	h := coins.NewCryptocoinHandler(cointype)
	if h == nil {
		data["result"] = ""
		return map[string]interface{}{
			"Status": "Error",
			"Tip":    "unsupported cointype",
			"Error":  "unsupported cointype",
			"Data":   data,
		}
	}

	v := h.GetDefaultFee()
	b, _ := json.Marshal(&v)

	data["result"] = string(b)
	return map[string]interface{}{
		"Status": "Success",
		"Tip":    "",
		"Error":  "",
		"Data":   data,
	}
}

func (this *Service) GetAddressBalance(address string, cointype string) map[string]interface{} {
	fmt.Printf("=====================call rpc GetAddressBalance, address = %v, cointype = %v ===========================\n", address, cointype)
	data := make(map[string]interface{})
	if cointype == "" || address == "" {
		data["result"] = ""
		return map[string]interface{}{
			"Status": "Error",
			"Tip":    "param error",
			"Error":  "param error",
			"Data":   data,
		}
	}

	h := coins.NewCryptocoinHandler(cointype)
	if h == nil {
		data["result"] = ""
		return map[string]interface{}{
			"Status": "Error",
			"Tip":    "unsupported cointype",
			"Error":  "unsupported cointype",
			"Data":   data,
		}
	}

	balance, err := h.GetAddressBalance(address, "")
	fmt.Printf("=====================call rpc GetAddressBalance, address = %v, cointype = %v,balance = %v, err = %v ===========================\n", address, cointype, balance, err)
	if err != nil {
		data["result"] = ""
		return map[string]interface{}{
			"Status": "Error",
			"Tip":    err.Error(),
			"Error":  err.Error(),
			"Data":   data,
		}
	}

	b, _ := json.Marshal(&balance)

	data["result"] = string(b)
	return map[string]interface{}{
		"Status": "Success",
		"Tip":    "",
		"Error":  "",
		"Data":   data,
	}
}

func (this *Service) IsToken(cointype string) map[string]interface{} {
	fmt.Printf("=====================call rpc IsToken, cointype = %v ===========================\n", cointype)
	data := make(map[string]interface{})
	if cointype == "" {
		data["result"] = ""
		return map[string]interface{}{
			"Status": "Error",
			"Tip":    "param error",
			"Error":  "param error",
			"Data":   data,
		}
	}

	h := coins.NewCryptocoinHandler(cointype)
	if h == nil {
		data["result"] = ""
		return map[string]interface{}{
			"Status": "Error",
			"Tip":    "unsupported cointype",
			"Error":  "unsupported cointype",
			"Data":   data,
		}
	}

	v := h.IsToken()

	data["result"] = v
	return map[string]interface{}{
		"Status": "Success",
		"Tip":    "",
		"Error":  "",
		"Data":   data,
	}
}

func FiltTransaction(blocknumber int64, from, receipient, contract, cointype string) ([]string, error) {
	h := coins.NewCryptocoinHandler(cointype)
	if h == nil {
		return nil, fmt.Errorf("unsupported cointype")
	}
	filter := types.Filter{
		From:       from,
		Receipient: receipient,
		Contract:   contract,
	}
	return h.FiltTransaction(uint64(blocknumber), filter)
}

func (this *Service) FiltTransaction(blocknumber int64, from, receipient, contract, cointype string) map[string]interface{} {
	txs, err := FiltTransaction(blocknumber, from, receipient, contract, cointype)
	return map[string](interface{}){
		"Transactions": txs,
		"Error":        err,
	}
}
