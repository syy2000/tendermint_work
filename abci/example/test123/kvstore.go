package test123

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	dbm "github.com/tendermint/tm-db"

	"strings"

	"github.com/tendermint/tendermint/abci/example/code"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/version"
)

var (
	stateKey        = []byte("stateKey")
	kvPairPrefixKey = []byte("kvPairKey:")

	ProtocolVersion uint64 = 0x1
)

type State struct {
	db1, db2, db3 dbm.DB
	Size          int64  `json:"size"`
	Height        int64  `json:"height"`
	AppHash       []byte `json:"app_hash"`
}

func loadState(dbDir string) State {
	var state State
	db1, err1 := dbm.NewGoLevelDB("statedb1", dbDir)
	if err1 != nil {
		panic(err1)
	}
	state.db1 = db1
	db2, err2 := dbm.NewGoLevelDB("statedb2", dbDir)
	if err2 != nil {
		panic(err2)
	}
	state.db2 = db2
	db3, err3 := dbm.NewGoLevelDB("statedb3", dbDir)
	if err3 != nil {
		panic(err3)
	}
	state.db3 = db3
	stateBytes, err := db1.Get(stateKey)
	if err != nil {
		panic(err)
	}
	if len(stateBytes) == 0 {
		return state
	}
	err = json.Unmarshal(stateBytes, &state)
	if err != nil {
		panic(err)
	}
	return state
}

func saveState(state State) {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	err = state.db1.Set(stateKey, stateBytes)
	if err != nil {
		panic(err)
	}
}

func prefixKey(key []byte) []byte {
	return append(kvPairPrefixKey, key...)
}

//---------------------------------------------------

var _ types.Application = (*Application)(nil)

type Application struct {
	types.BaseApplication

	state        State
	RetainBlocks int64 // blocks to retain after commit (via ResponseCommit.RetainHeight)
}

func NewApplication(dbDir string) *Application {
	state := loadState(dbDir)
	return &Application{state: state}
}

func (app *Application) Info(req types.RequestInfo) (resInfo types.ResponseInfo) {
	return types.ResponseInfo{
		Data:             fmt.Sprintf("{\"size\":%v}", app.state.Size),
		Version:          version.ABCIVersion,
		AppVersion:       ProtocolVersion,
		LastBlockHeight:  app.state.Height,
		LastBlockAppHash: app.state.AppHash,
	}
}

// tx is either "key=value" or just arbitrary bytes
func (app *Application) DeliverTx(req types.RequestDeliverTx) types.ResponseDeliverTx {
	var data map[string]string
	err := json.Unmarshal([]byte(req.Tx), &data)
	if err != nil {
		panic(err)
	}
	m := data["TxObject"]
	fields := strings.Split(m, "-")
	data["id"] = fields[0]
	data["attr"] = fields[1]
	operation := data["TxOp"]
	if operation == "write" {
		key1 := data["TxObject"]
		value := data["TxValue"]
		err1 := app.state.db1.Set(prefixKey([]byte(key1)), []byte(value))
		if err1 != nil {
			panic(err1)
		}
		key2 := data["id"]
		value2 := data["attr"]
		err2 := app.state.db2.Set(prefixKey([]byte(key2)), []byte(value2))
		if err2 != nil {
			panic(err2)
		}
		key3 := data["attr"]
		value3 := data["TxValue"]
		err3 := app.state.db3.Set(prefixKey([]byte(key3)), []byte(value3))
		if err3 != nil {
			panic(err3)
		}
	}
	app.state.Size++

	events := []types.Event{
		{
			Type: "app",
			Attributes: []types.EventAttribute{
				{Key: []byte("creator"), Value: []byte("Cosmoshi Netowoko"), Index: true},
				{Key: []byte("key"), Value: []byte(req.String()), Index: true},
				{Key: []byte("index_key"), Value: []byte("index is working"), Index: true},
				{Key: []byte("noindex_key"), Value: []byte("index is working"), Index: false},
			},
		},
	}

	return types.ResponseDeliverTx{Code: code.CodeTypeOK, Events: events}
}

func (app *Application) CheckTx(req types.RequestCheckTx) types.ResponseCheckTx {
	return types.ResponseCheckTx{Code: code.CodeTypeOK, GasWanted: 1}
}

func (app *Application) Commit() types.ResponseCommit {
	// Using a memdb - just return the big endian size of the db
	appHash := make([]byte, 8)
	binary.PutVarint(appHash, app.state.Size)
	app.state.AppHash = appHash
	app.state.Height++
	saveState(app.state)

	resp := types.ResponseCommit{Data: appHash}
	if app.RetainBlocks > 0 && app.state.Height >= app.RetainBlocks {
		resp.RetainHeight = app.state.Height - app.RetainBlocks + 1
	}
	return resp
}

// Returns an associated value or nil if missing.
func (app *Application) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
	if reqQuery.Prove {
		value, err := app.state.db1.Get(prefixKey(reqQuery.Data))
		if err != nil {
			panic(err)
		}
		if value == nil {
			resQuery.Log = "does not exist"
		} else {
			resQuery.Log = "exists"
		}
		resQuery.Index = -1 // TODO make Proof return index
		resQuery.Key = reqQuery.Data
		resQuery.Value = value
		resQuery.Height = app.state.Height

		return
	}

	resQuery.Key = reqQuery.Data
	value, err := app.state.db1.Get(prefixKey(reqQuery.Data))
	if err != nil {
		panic(err)
	}
	if value == nil {
		resQuery.Log = "does not exist"
	} else {
		resQuery.Log = "exists"
	}
	resQuery.Value = value
	resQuery.Height = app.state.Height

	return resQuery
}
