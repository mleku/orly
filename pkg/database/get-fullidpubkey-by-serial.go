package database

import (
	"bytes"
	"github.com/dgraph-io/badger/v4"
	"orly.dev/pkg/database/indexes"
	"orly.dev/pkg/database/indexes/types"
	"orly.dev/pkg/interfaces/store"
	"orly.dev/pkg/utils/chk"
)

func (d *D) GetFullIdPubkeyBySerial(ser *types.Uint40) (
	fidpk *store.IdPkTs, err error,
) {
	if err = d.View(
		func(txn *badger.Txn) (err error) {
			buf := new(bytes.Buffer)
			if err = indexes.FullIdPubkeyEnc(
				ser, nil, nil, nil,
			).MarshalWrite(buf); chk.E(err) {
				return
			}
			prf := buf.Bytes()
			it := txn.NewIterator(
				badger.IteratorOptions{
					Prefix: prf,
				},
			)
			defer it.Close()
			it.Seek(prf)
			if it.Valid() {
				item := it.Item()
				key := item.KeyCopy(nil)
				ser, fid, p, ca := indexes.FullIdPubkeyVars()
				buf2 := bytes.NewBuffer(key)
				if err = indexes.FullIdPubkeyDec(
					ser, fid, p, ca,
				).UnmarshalRead(buf2); chk.E(err) {
					return
				}
				idpkts := store.IdPkTs{
					Id:  fid.Bytes(),
					Pub: p.Bytes(),
					Ts:  int64(ca.Get()),
					Ser: ser.Get(),
				}
				fidpk = &idpkts
			}
			return
		},
	); chk.E(err) {
		return
	}
	return
}
