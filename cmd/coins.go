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
	"os"
	"net"
	"math/big"

	"os/signal"
	"strconv"
	"strings"
	rpc "github.com/fsn-dev/cryptoCoins/internal/rpcservice"
	"github.com/fsn-dev/cryptoCoins/coins"
	"gopkg.in/urfave/cli.v1"
	"encoding/json"
	"github.com/fsn-dev/cryptoCoins/types"
	cryptocoinsconfig "github.com/fsn-dev/cryptoCoins/config"
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
	rpcport   int
	app       = cli.NewApp()
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
	rpc_port  int
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

func GetAddr(pubkey string,cointype string) (string,string,error) {
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

func (this *Service) GetAddr(pubkey string,cointype string) map[string]interface{} {
    fmt.Printf("=====================call rpc GetAddr, pubkey = %v, cointype = %v ===========================\n",pubkey,cointype)
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

    ret, tip, err := GetAddr(pubkey,cointype)
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

type TransactionInfo struct {
    FromAddress string
    TxOutPuts []types.TxOutput
    JsonString string
    Confirmed bool
    Fee types.Value
}

func GetTransactionInfo(txhash string,cointype string) (*TransactionInfo,string,error) {
    h := coins.NewCryptocoinHandler(cointype)
    if h == nil {
	return nil,"unsupported cointype",fmt.Errorf("unsupported cointype")
    }

    from,txout,jsonstr,confir,fee,err := h.GetTransactionInfo(txhash)
    ti := &TransactionInfo{FromAddress:from,TxOutPuts:txout,JsonString:jsonstr,Confirmed:confir,Fee:fee}
    return ti,"",err
}

func (this *Service) GetTransactionInfo(txhash string,cointype string) map[string]interface{} {
    fmt.Printf("=====================call rpc GetTransactionInfo, txhash = %v, cointype = %v ===========================\n",txhash,cointype)
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

    ret, tip, err := GetTransactionInfo(txhash,cointype)
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
    Tx string
    TxHash []string
}

func BuildUnsignedTransaction(fromaddr string,pubkey string,toaddr string,amount string,cointype string) (*UnsignTx,string,error) {
    h := coins.NewCryptocoinHandler(cointype)
    if h == nil {
	return nil,"unsupported cointype",fmt.Errorf("unsupported cointype")
    }

    value, ok := new(big.Int).SetString(amount, 10)
    if ok == false {
	return nil,"get value error",fmt.Errorf("get value error")
    }

    tx,txhash,err := h.BuildUnsignedTransaction(fromaddr,pubkey,toaddr,value,"")
    b, _ := json.Marshal(tx)
    ut := &UnsignTx{Tx:string(b),TxHash:txhash}
    return ut,"",err
}

func (this *Service) BuildUnsignedTransaction(fromaddr string,pubkey string,toaddr string,amount string,cointype string) map[string]interface{} {
    fmt.Printf("=====================call rpc BuildUnsignedTransaction, fromaddr = %v, pubkey = %v, toaddr = %v, amount = %v, cointype = %v ===========================\n",fromaddr,pubkey,toaddr,amount,cointype)
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

    ret, tip, err := BuildUnsignedTransaction(fromaddr,pubkey,toaddr,amount,cointype)
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

func MakeSignedTransaction(txjson string,rsv string,cointype string) (string,string,error) {
    h := coins.NewCryptocoinHandler(cointype)
    if h == nil {
	return "","unsupported cointype",fmt.Errorf("unsupported cointype")
    }

    var tx interface{}
    json.Unmarshal([]byte(txjson), &tx)
    rsvs := make([]string,0)
    rsvs = append(rsvs,rsv)
    signtx,err := h.MakeSignedTransaction(rsvs,tx)
    b, _ := json.Marshal(signtx)
    return string(b),"",err
}

func (this *Service) MakeSignedTransaction(tx string,rsv string,cointype string) map[string]interface{} {
    fmt.Printf("=====================call rpc MakeSignedTransaction, tx = %v, rsv = %v, cointype = %v ===========================\n",tx,rsv,cointype)
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

    ret, tip, err := MakeSignedTransaction(tx,rsv,cointype)
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

func SubmitTransaction(signtx string,cointype string) (string,string,error) {
    h := coins.NewCryptocoinHandler(cointype)
    if h == nil {
	return "","unsupported cointype",fmt.Errorf("unsupported cointype")
    }

    var tx interface{}
    json.Unmarshal([]byte(signtx), &tx)

    txhash,err := h.SubmitTransaction(tx)
    return txhash,"",err
}

func (this *Service) SubmitTransaction(signtx string,cointype string) map[string]interface{} {
    fmt.Printf("=====================call rpc ,SubmitTransaction, signtx = %v, cointype = %v ===========================\n",signtx,cointype)
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

    ret, tip, err := SubmitTransaction(signtx,cointype)
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
    fmt.Printf("=====================call rpc GetDefaultFee, cointype = %v ===========================\n",cointype)
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

func (this *Service) GetAddressBalance(address string,cointype string) map[string]interface{} {
    fmt.Printf("=====================call rpc GetAddressBalance, address = %v, cointype = %v ===========================\n",address,cointype)
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

    balance,err := h.GetAddressBalance(address,"")
    fmt.Printf("=====================call rpc GetAddressBalance, address = %v, cointype = %v,balance = %v, err = %v ===========================\n",address,cointype,balance,err)
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
    fmt.Printf("=====================call rpc IsToken, cointype = %v ===========================\n",cointype)
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

