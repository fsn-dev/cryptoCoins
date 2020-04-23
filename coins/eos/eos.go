package eos

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"runtime"
	"path/filepath"
	"os/user"
	"os"

	"github.com/astaxie/beego/logs"
	"github.com/fsn-dev/cryptoCoins/coins/config"
	"github.com/fsn-dev/cryptoCoins/coins/rpcutils"
	"github.com/fsn-dev/cryptoCoins/tools/ethdb"
)

var (
	lock sync.Mutex
)

// DefaultDataDir is the default data directory to use for the databases and other
// persistence requirements.
func DefaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "coins-service")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "coins-service")
		} else {
			return filepath.Join(home, ".coins-service")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

//eos_init---> eos account
//key: crypto.Keccak256Hash([]byte("eossettings"))
//value: pubkey+eos account
func GetEosDbDir() string {
	dir := DefaultDataDir()
	dir += "/dcrmdata/eosdb"
	return dir
}

var EOSRegExpmap map[string]string = map[string]string{
	"EOSDCRM": "^d[1-5a-z]{32,33}$",
	//"EOS":"^(d[1-5a-z]{32,33})|([1-5a-z]{12})$",
	"EOS":        "^([a-z\\d\\.]+)$",
	"EOS_NORMAL": "^([1-5a-z]{12})$",
}

func CreateRealEosAccount(accountName string, ownerkey string, activekey string) error {
	fmt.Println("==========create eos account Start!!,account name = %s,ownerkey = %s,activekey = %s, ==============", accountName, ownerkey, activekey)
	match, _ := regexp.MatchString(EOSRegExpmap["EOS_NORMAL"], accountName)
	//av := coins.NewAddressValidator("EOS_NORMAL")
	//if !av.IsValidAddress(accountName) {
	if !match {
		return errors.New("eos account name format error")
	}
	owner, e1 := HexToPubKey(ownerkey)
	active, e2 := HexToPubKey(activekey)
	if e1 != nil || e2 != nil {
		fmt.Println("==========create eos account ==========,ownerkey error = %+v,active error = %+v", e1, e2)
		return errors.New("cannot convert to eos pubkey format.")
	}
	fmt.Println("==========create eos account,owner = %s,active = %s,==============", owner, active)
	///*
	ojbk, err := CreateNewAccount(CREATOR_ACCOUNT, CREATOR_PRIVKEY, accountName, owner.String(), active.String(), InitialRam)
	if ojbk == false || err != nil {
		fmt.Println("create eos account failed,error = %+v", err)
		return errors.New("create eos account failed")
	}
	ojbk2, err := DelegateBW(CREATOR_ACCOUNT, CREATOR_PRIVKEY, accountName, InitialCPU, InitialStakeNet, true)
	if ojbk2 == false || err != nil {
		fmt.Println("delegate cpu and net failed,error = %+v", err)
		return errors.New("delegate cpu and net failed")
	}

	return err
	//*/
	//	return nil
}

var trytimes = 50

func CheckRealEosAccount(accountName, ownerkey, activekey string) (ok bool) {
	ok = false
	defer func() {
		if r := recover(); r != nil {
			logs.Debug("check eos account", "error", r)
		}
	}()
	logs.Debug("==========check eos account Start!!==========")
	// 1. check if account exists
	api := "v1/chain/get_account"
	data := `{"account_name":"` + accountName + `"}`
	var ret string
	info := new(AccountInfo)
	var err error
	for i := 0; i < trytimes; i++ {
		ret = rpcutils.DoCurlRequest(config.ApiGateways.EosGateway.Nodeos, api, data)
		logs.Debug("========check eos account========", "ret", ret)
		err = json.Unmarshal([]byte(ret), info)
		if err != nil {
			logs.Debug("========check eos account========", "decode error", err)
			time.Sleep(time.Duration(1) * time.Second)
			continue
		}
		break
	}
	if err != nil {
		logs.Debug("decode account info error")
		return false
	}
	if info.Error != nil || info.AccountName != accountName {
		logs.Debug("check real eos account error", "error", info.Error)
		return false
	}
	// 2. check owner key
	// 3. check active key
	// 4. check no other keys authorized

	owner, e1 := HexToPubKey(ownerkey) //EOS8JXJf7nuBEs8dZ8Pc5NpS8BJJLt6bMAmthWHE8CSqzX4VEFKtq
	active, e2 := HexToPubKey(activekey)
	if e1 != nil || e2 != nil {
		logs.Debug("public key error", "owner key error", e1, "active key error", e2)
		return false
	}

	perm := info.Perms
	objPerm := Permissions([]Permission{
		{PermName: "owner", Parent: "", RequiredAuth: Auth{Threshold: 1, Keys: []Key{{Key: owner.String(), Weight: 1}}}},
		{PermName: "active", Parent: "owner", RequiredAuth: Auth{Threshold: 1, Keys: []Key{{Key: active.String(), Weight: 1}}}},
	})
	sort.Sort(objPerm)
	sort.Sort(perm)
	if reflect.DeepEqual(perm, objPerm) == false {
		logs.Debug("account permissions not match", "have", perm, "required", objPerm)
		return false
	}
	// 5. enough ram cpu net
	if info.RamQuato-info.RamUsage < InitialRam/2 {
		logs.Debug("account ram is too low")
		return false
	}
	if int64(info.CpuLimit.Max) < InitialCPU*5 {
		logs.Debug("account cpu is too low")
		return false
	}
	if int64(info.NetLimit.Max) < InitialStakeNet*5 {
		logs.Debug("account net bandwidth is too low")
		return false
	}
	logs.Debug("==========check eos account Success!!==========")
	return true
}

type AccountInfo struct {
	AccountName string                 `json:"account_name"`
	RamQuato    uint32                 `json:"ram_quato"`
	NetWeight   uint32                 `json:"net_weight"`
	CpuWeight   uint32                 `json:"cpu_weight"`
	NetLimit    Limit                  `json:"net_limit"`
	CpuLimit    Limit                  `json:"cpu_limit"`
	RamUsage    uint32                 `json:"ram_usage"`
	Perms       Permissions            `json:"permissions"`
	Error       map[string]interface{} `json:"error"`
}

type Permissions []Permission

func (p Permissions) Len() int {
	return len([]Permission(p))
}

func (p Permissions) Less(i, j int) bool {
	if []Permission(p)[i].PermName == "owner" {
		return true
	}
	return false
}

func (p Permissions) Swap(i, j int) {
	tmp := []Permission(p)[i]
	[]Permission(p)[i] = []Permission(p)[j]
	[]Permission(p)[j] = tmp
}

type Permission struct {
	PermName     string `json:"perm_name"`
	Parent       string `json:"parent"`
	RequiredAuth Auth   `json:"required_auth"`
}

type Auth struct {
	Threshold int   `json:"threshold"`
	Keys      []Key `json:"keys"`
}

type Key struct {
	Key    string `json:"key"`
	Weight int    `json:"weight"`
}

type Limit struct {
	Used      int64 `used`
	Available int64 `available`
	Max       int64 `max`
}

func GetEosAccount() (acct, owner, active string) {
	lock.Lock()
	dir := GetEosDbDir()
	db, err := ethdb.NewLDBDatabase(dir, 0, 0)
	if err != nil {
		logs.Debug("==============open db fail.============")
		lock.Unlock()
		return "", "", ""
	}

	var data string
	var b bytes.Buffer
	b.WriteString("")
	b.WriteByte(0)
	b.WriteString("")
	iter := db.NewIterator()
	for iter.Next() {
		key := string(iter.Key())
		value := string(iter.Value())
		if strings.EqualFold(key, string([]byte("eossettings"))) {
			data = value
			break
		}
	}
	iter.Release()
	if data == "" {
		fmt.Println("===============GetEosAccount,get data fail.==================")
		db.Close()
		lock.Unlock()
		return "", "", ""
	}

	datas := strings.Split(string(data), ":")
	if len(datas) == 5 && datas[0] == "EOS_INITIATE" {
		db.Close()
		lock.Unlock()
		return datas[1], datas[2], datas[3]
	}
	db.Close()
	lock.Unlock()
	return "", "", ""
}
