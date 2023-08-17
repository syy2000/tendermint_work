package scoin

import "encoding/json"

func CreateInsertTx(pubkey string, money int32) []byte {
	insert := Insert{
		Flag:   1,
		Pubkey: pubkey,
		Money:  money,
	}
	if o, err := json.Marshal(insert); err != nil {
		panic(err)
	} else {
		return Filt(o)
	}
}

func CreateTransferTx(froms, tos []string, money int32) []byte {
	buy := Buy{
		Flag:  2,
		From:  froms,
		To:    tos,
		Sigs:  nil,
		Money: money,
	}
	if o, err := json.Marshal(buy); err != nil {
		panic(err)
	} else {
		return Filt(o)
	}
}
