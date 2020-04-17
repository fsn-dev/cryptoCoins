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

	"os/signal"
	"strconv"
	"strings"
	rpc "github.com/fsn-dev/cryptoCoins/rpcservice"
	"gopkg.in/urfave/cli.v1"
)

func main() {

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func StartCoins(c *cli.Context) {
	RpcInit(rpcport)
	select {} // note for server, or for client
}

var (
	rpcport   int
	port      int
	app       = cli.NewApp()
)

func init() {
	app.Usage = "Crypto Coins Service"
	app.Version = "0.0"
	app.Action = StartCoins
	app.Flags = []cli.Flag{
		cli.IntFlag{Name: "rpcport", Value: 0, Usage: "listen port", Destination: &rpcport},
		cli.IntFlag{Name: "port", Value: 0, Usage: "listen port", Destination: &port},
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

func (this *Service) GetAddr(pubkey string,cointype string) map[string]interface{} {
    fmt.Printf("=====================call rpc GetAddr, pubkey = %v, cointype = %v ===========================\n",pubkey,cointype)
	/*data := make(map[string]interface{})
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
	*/

	return nil
}

