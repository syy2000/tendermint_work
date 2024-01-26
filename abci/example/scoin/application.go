package scoin

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/tendermint/tendermint/version"
	dbm "github.com/tendermint/tm-db"

	codetype "github.com/tendermint/tendermint/abci/example/code"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/tendermint/tendermint/crypto/ed25519"

	abcitypes "github.com/tendermint/tendermint/abci/types"
	pc "github.com/tendermint/tendermint/proto/tendermint/crypto"
	"github.com/tendermint/tendermint/types"
)

func verify(msg, pubkey, sig []byte) bool {
	if len(msg) > 0 {
		return true
	}
	u1, err1 := hex.DecodeString(string(pubkey))
	u2, err2 := hex.DecodeString(string(sig))
	if err1 != nil || err2 != nil || len(u1) != 32 {
		return false
	}
	pbk := ed25519.PubKey(u1)
	return pbk.VerifySignature(msg, u2)
}

type Application struct {
	abcitypes.BaseApplication

	dbcnn dbm.DB

	state        State
	RetainBlocks int64 // blocks to retain after commit (via ResponseCommit.RetainHeight)

	// validator set
	ValUpdates         []abcitypes.ValidatorUpdate
	valAddrToPubKeyMap map[string]pc.PublicKey

	logger log.Logger
}

func NewApplication(dbDir string) *Application {

	var state State

	if dbDir == "" {
		statedb := dbm.NewMemDB()
		state = loadState(statedb)
	} else {
		statedb, err := dbm.NewGoLevelDB("concurrency coin", dbDir)
		if err != nil {
			panic(err)
		}
		state = loadState(statedb)
	}

	return &Application{
		dbcnn:              state.db,
		valAddrToPubKeyMap: make(map[string]pc.PublicKey),
		logger:             log.NewNopLogger(),
		state:              state,
	}
}

func (app *Application) GetDB() dbm.DB {
	return app.state.db
}

type Insert struct {
	Flag   int8
	Pubkey string
	Money  int32
}

type Buy struct {
	Flag  int8
	From  []string
	To    []string
	Sigs  []string
	Money int32
}

// Return application info
func (app *Application) Info(req abcitypes.RequestInfo) abcitypes.ResponseInfo {
	return abcitypes.ResponseInfo{
		Data:             fmt.Sprintf("{\"size\":%v}", app.state.Size),
		Version:          version.ABCISemVer,
		AppVersion:       ProtocolVersion,
		LastBlockHeight:  app.state.Height,
		LastBlockAppHash: app.state.AppHash,
	}
}

// Query for state
func (app *Application) Query(reqQuery abcitypes.RequestQuery) (resQuery abcitypes.ResponseQuery) {
	if reqQuery.Prove {
		value, err := app.state.db.Get(prefixKey(reqQuery.Data))
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
		resQuery.Value = []byte(fmt.Sprintf("%d", B2I(value)))
		resQuery.Info = fmt.Sprintf("%d", B2I(value))
		resQuery.Height = app.state.Height

		return
	}

	resQuery.Key = reqQuery.Data
	value, err := app.state.db.Get(prefixKey(reqQuery.Data))
	if err != nil {
		panic(err)
	}
	if value == nil {
		resQuery.Log = "does not exist"
	} else {
		resQuery.Log = "exists"
	}
	resQuery.Value = []byte(fmt.Sprintf("%d", B2I(value)))
	resQuery.Info = fmt.Sprintf("%d", B2I(value))
	resQuery.Height = app.state.Height

	return resQuery
}

//------------- mempool connection-----------------------------------------------------------

// Validate a tx for the mempool
func (app *Application) CheckTx(req abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	if isValidatorTx(req.Tx.OriginTx) {
		return abcitypes.ResponseCheckTx{Code: codetype.CodeTypeOK, GasWanted: 1, Log: string(req.Tx.OriginTx)}
	}
	var insert Insert
	var buy Buy
	err := json.Unmarshal(UnFilt(req.Tx.OriginTx), &insert)
	if err != nil {
		return abcitypes.ResponseCheckTx{Code: codetype.CodeTypeEncodingError, GasWanted: 0, Log: string(req.Tx.OriginTx), Info: "Wrong json"}
	}
	if insert.Flag == int8(1) {
		if len(insert.Pubkey) == 0 {
			return abcitypes.ResponseCheckTx{Code: codetype.CodeTypeBadNonce, GasWanted: 0, Log: string(req.Tx.OriginTx), Info: "nil insert"}
		}
		return abcitypes.ResponseCheckTx{Code: codetype.CodeTypeOK, GasWanted: 1, Log: string(req.Tx.OriginTx)}
	} else if insert.Flag == int8(2) || insert.Flag == int8(3) {

		if err := json.Unmarshal(UnFilt(req.Tx.OriginTx), &buy); err != nil {
			return abcitypes.ResponseCheckTx{Code: codetype.CodeTypeBadNonce, GasWanted: 0, Log: string(req.Tx.OriginTx), Info: "wrong json"}
		}
		if len(buy.From) == 0 || len(buy.To) == 0 {
			return abcitypes.ResponseCheckTx{Code: codetype.CodeTypeBadNonce, GasWanted: 0, Log: string(req.Tx.OriginTx), Info: "nil transfer"}
		}
		if buy.Money <= 0 {
			return abcitypes.ResponseCheckTx{Code: codetype.CodeTypeBadNonce, GasWanted: 0, Log: string(req.Tx.OriginTx), Info: "bad transfer"}
		}

		return abcitypes.ResponseCheckTx{Code: codetype.CodeTypeOK, GasWanted: 1, Log: string(req.Tx.OriginTx)}
	} else {
		return abcitypes.ResponseCheckTx{Code: codetype.CodeTypeEncodingError, GasWanted: 0, Log: string(req.Tx.OriginTx), Info: "Unknown Service"}
	}

}

// -------------Consensus Connection------------------------------------------------------------

func (app *Application) InitChain(req abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	for _, v := range req.Validators {
		r := app.updateValidator(v)
		if r.IsErr() {
			app.logger.Error("Error updating validators", "r", r)
		}
	}
	return abcitypes.ResponseInitChain{}
}

func (app *Application) BeginBlock(req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	// reset valset changes
	app.ValUpdates = make([]abcitypes.ValidatorUpdate, 0)

	// Punish validators who committed equivocation.
	for _, ev := range req.ByzantineValidators {
		if ev.Type == abcitypes.EvidenceType_DUPLICATE_VOTE {
			addr := string(ev.Validator.Address)
			if pubKey, ok := app.valAddrToPubKeyMap[addr]; ok {
				app.updateValidator(abcitypes.ValidatorUpdate{
					PubKey: pubKey,
					Power:  ev.Validator.Power - 1,
				})
				app.logger.Info("Decreased val power by 1 because of the equivocation",
					"val", addr)
			} else {
				app.logger.Error("Wanted to punish val, but can't find it",
					"val", addr)
			}
		}
	}

	return abcitypes.ResponseBeginBlock{}
}

func (app *Application) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	app.Execute(req.Tx.OriginTx)
	app.state.Size++
	return abcitypes.ResponseDeliverTx{Code: abcitypes.CodeTypeOK}
}

// Signals the end of a block, returns changes to the validator set
func (app *Application) EndBlock(abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	return abcitypes.ResponseEndBlock{ValidatorUpdates: app.ValUpdates}
}

func (app *Application) Commit() abcitypes.ResponseCommit {
	app.state.Height++

	appHash := make([]byte, 8)
	binary.PutVarint(appHash, app.state.Size)
	app.state.AppHash = appHash

	// empty out the set of transactions to remove via rechecktx
	saveState(app.state)

	resp := abcitypes.ResponseCommit{Data: app.state.AppHash}
	if app.RetainBlocks > 0 && app.state.Height >= app.RetainBlocks {
		resp.RetainHeight = app.state.Height - app.RetainBlocks + 1
	}
	return resp
}

// --------------------State Sync Connection---------------------------------------
func (app *Application) ListSnapshots(abcitypes.RequestListSnapshots) abcitypes.ResponseListSnapshots { // List available snapshots
	return abcitypes.ResponseListSnapshots{}
}
func (app *Application) OfferSnapshot(abcitypes.RequestOfferSnapshot) abcitypes.ResponseOfferSnapshot { // Offer a snapshot to the application
	return abcitypes.ResponseOfferSnapshot{}
}
func (app *Application) LoadSnapshotChunk(abcitypes.RequestLoadSnapshotChunk) abcitypes.ResponseLoadSnapshotChunk { // Load a snapshot chunk
	return abcitypes.ResponseLoadSnapshotChunk{}
}
func (app *Application) ApplySnapshotChunk(abcitypes.RequestApplySnapshotChunk) abcitypes.ResponseApplySnapshotChunk { // Apply a shapshot chunk
	return abcitypes.ResponseApplySnapshotChunk{}
}

// --------------execute-----------------------
func (app *Application)  Execute(tx []byte) (abcitypes.ResponseDeliverTx, bool) {
	if isValidatorTx(tx) {
		// update validators in the merkle tree
		// and in app.ValUpdates
		resp := app.execValidatorTx(tx)
		return resp, resp.IsOK()
	}

	//fmt.Println("in coin")
	var insert Insert
	var buy Buy

	err := json.Unmarshal(UnFilt(tx), &insert)

	if err != nil {
		return abcitypes.ResponseDeliverTx{Code: codetype.CodeTypeBadNonce}, false
	}
	if insert.Flag == int8(1) {
		fmt.Printf("%d\n", insert.Flag)
		key := prefixKey([]byte(insert.Pubkey))
		value := I2B(insert.Money)
		out, err := app.dbcnn.Get(key)
		//fmt.Println("I'm sure it is ok")
		if out == nil && err == nil {
			if err := app.dbcnn.Set(key, value); err != nil {
				return abcitypes.ResponseDeliverTx{Code: codetype.CodeTypeBadNonce}, false
			}
			return abcitypes.ResponseDeliverTx{Code: codetype.CodeTypeOK}, true
		} else {
			return abcitypes.ResponseDeliverTx{Code: codetype.CodeTypeBadNonce}, false
		}
	} else if insert.Flag == int8(2) || insert.Flag == int8(3) {
		fmt.Printf("%d\n", insert.Flag)
		json.Unmarshal(UnFilt(tx), &buy)

		numfrom, numto := len(buy.From), len(buy.To)

		if numfrom != len(buy.Sigs) {
			return abcitypes.ResponseDeliverTx{Code: codetype.CodeTypeBadNonce}, false
		}

		fromkeys := make([]ed25519.PubKey, numfrom)
		tokeys := make([]ed25519.PubKey, numto)

		msg := []byte("ok")

		for i := 0; i < numfrom; i++ {
			fromkeys[i] = []byte(buy.From[i])
			//fmt.Println(len(fromkeys[i]))
			if !verify(msg, []byte(buy.From[i]), []byte(buy.Sigs[i])) {
				return abcitypes.ResponseDeliverTx{Code: codetype.CodeTypeBadNonce}, false
			}
		}

		for i := 0; i < numto; i++ {
			tokeys[i] = []byte(buy.To[i])
		}

		frompay := int32(numto) * buy.Money
		toget := int32(numfrom) * buy.Money

		froms := make([]int32, numfrom)
		tos := make([]int32, numto)

		for i := 0; i < numfrom; i++ {
			if m, err := app.dbcnn.Get(prefixKey(fromkeys[i])); err == nil && m != nil {
				froms[i] = B2I(m)
				if froms[i] < frompay {
					return abcitypes.ResponseDeliverTx{Code: codetype.CodeTypeBadNonce}, false
				} else {
					froms[i] -= frompay
				}
			} else {
				return abcitypes.ResponseDeliverTx{Code: codetype.CodeTypeBadNonce}, false
			}
		}

		for i := 0; i < numto; i++ {
			if m, err := app.dbcnn.Get(prefixKey(tokeys[i])); err == nil && m != nil {
				tos[i] = B2I(m) + toget
			} else {
				return abcitypes.ResponseDeliverTx{Code: codetype.CodeTypeBadNonce}, false
			}
		}

		for i := 0; i < numfrom; i++ {
			if err := app.dbcnn.Set(prefixKey(fromkeys[i]), I2B(froms[i])); err != nil {
				panic("db error : " + err.Error())
			}
		}

		for i := 0; i < numto; i++ {
			if err := app.dbcnn.Set(prefixKey(tokeys[i]), I2B(tos[i])); err != nil {
				panic("db error : " + err.Error())
			}
		}
		return abcitypes.ResponseDeliverTx{Code: codetype.CodeTypeOK}, true
	} else {
		return abcitypes.ResponseDeliverTx{Code: codetype.CodeTypeBadNonce}, false
	}
}

func RWAnalyse(stx types.Tx) (op, address []string, err error) {
	tx := stx.OriginTx
	if isValidatorTx(tx) {
		op = []string{"write"}
		address = []string{ValidatorSetChangePrefix}
		err = nil
		return
	}

	var insert Insert
	var buy Buy
	err = json.Unmarshal(UnFilt(tx), &insert)
	if err != nil {
		return
	}
	if insert.Flag == int8(1) {
		if len(insert.Pubkey) == 0 {
			err = errInsertNoKey
			return
		}
		op = []string{"write"}
		address = []string{insert.Pubkey}
		return
	} else if insert.Flag == int8(2) {
		err = json.Unmarshal(UnFilt(tx), &buy)
		if err != nil {
			return
		}
		for i := 0; i < len(buy.From); i++ {
			op = append(op, "write")
			address = append(address, buy.From[i])
		}
		for i := 0; i < len(buy.To); i++ {
			op = append(op, "write")
			address = append(address, buy.To[i])
		}
		return
	} else if insert.Flag == int8(3) {
		err = json.Unmarshal(UnFilt(tx), &buy)
		if err != nil {
			return
		}
		for i := 0; i < len(buy.From); i++ {
			for _, u := range TestList {
				op = append(op, "write")
				address = append(address, string(u)+buy.From[i])
			}
		}
		for i := 0; i < len(buy.To); i++ {
			for _, u := range TestList {
				op = append(op, "write")
				address = append(address, string(u)+buy.To[i])
			}
		}
		return
	} else {
		err = errUnMarshalableTx
		return
	}
}

var TestList = "0123456789"
