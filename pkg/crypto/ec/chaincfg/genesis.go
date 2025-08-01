package chaincfg

import (
	"orly.dev/pkg/crypto/ec/chainhash"
	"orly.dev/pkg/crypto/ec/wire"
	"time"
)

var (
	// genesisCoinbaseTx is the coinbase transaction for the genesis blocks for
	// the main network, regression test network, and test network (version 3).
	genesisCoinbaseTx = wire.MsgTx{
		Version: 1,
		TxIn: []*wire.TxIn{
			{
				PreviousOutPoint: wire.OutPoint{
					Hash:  chainhash.Hash{},
					Index: 0xffffffff,
				},
				SignatureScript: []byte{
					0x04, 0xff, 0xff, 0x00, 0x1d, 0x01, 0x04,
					0x45, /* |.......E| */
					0x54, 0x68, 0x65, 0x20, 0x54, 0x69, 0x6d,
					0x65, /* |The Time| */
					0x73, 0x20, 0x30, 0x33, 0x2f, 0x4a, 0x61,
					0x6e, /* |s 03/Jan| */
					0x2f, 0x32, 0x30, 0x30, 0x39, 0x20, 0x43,
					0x68, /* |/2009 Ch| */
					0x61, 0x6e, 0x63, 0x65, 0x6c, 0x6c, 0x6f,
					0x72, /* |ancellor| */
					0x20, 0x6f, 0x6e, 0x20, 0x62, 0x72, 0x69,
					0x6e, /* | on brin| */
					0x6b, 0x20, 0x6f, 0x66, 0x20, 0x73, 0x65,
					0x63, /* |k of sec|*/
					0x6f, 0x6e, 0x64, 0x20, 0x62, 0x61, 0x69,
					0x6c, /* |ond bail| */
					0x6f, 0x75, 0x74, 0x20, 0x66, 0x6f, 0x72,
					0x20,                         /* |out for |*/
					0x62, 0x61, 0x6e, 0x6b, 0x73, /* |banks| */
				},
				Sequence: 0xffffffff,
			},
		},
		TxOut: []*wire.TxOut{
			{
				Value: 0x12a05f200,
				PkScript: []byte{
					0x41, 0x04, 0x67, 0x8a, 0xfd, 0xb0, 0xfe,
					0x55, /* |A.g....U| */
					0x48, 0x27, 0x19, 0x67, 0xf1, 0xa6, 0x71,
					0x30, /* |H'.g..q0| */
					0xb7, 0x10, 0x5c, 0xd6, 0xa8, 0x28, 0xe0,
					0x39, /* |..\..(.9| */
					0x09, 0xa6, 0x79, 0x62, 0xe0, 0xea, 0x1f,
					0x61, /* |..yb...a| */
					0xde, 0xb6, 0x49, 0xf6, 0xbc, 0x3f, 0x4c,
					0xef, /* |..I..?L.| */
					0x38, 0xc4, 0xf3, 0x55, 0x04, 0xe5, 0x1e,
					0xc1, /* |8..U....| */
					0x12, 0xde, 0x5c, 0x38, 0x4d, 0xf7, 0xba,
					0x0b, /* |..\8M...| */
					0x8d, 0x57, 0x8a, 0x4c, 0x70, 0x2b, 0x6b,
					0xf1,             /* |.W.Lp+k.| */
					0x1d, 0x5f, 0xac, /* |._.| */
				},
			},
		},
		LockTime: 0,
	}
	// genesisHash is the hash of the first block in the block chain for the main
	// network (genesis block).
	genesisHash = chainhash.Hash(
		[chainhash.HashSize]byte{
			// Make go vet happy.
			0x6f, 0xe2, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72,
			0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
			0x93, 0x1e, 0x83, 0x65, 0xe1, 0x5a, 0x08, 0x9c,
			0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
		},
	)
	// genesisMerkleRoot is the hash of the first transaction in the genesis block
	// for the main network.
	genesisMerkleRoot = chainhash.Hash(
		[chainhash.HashSize]byte{
			// Make go vet happy.
			0x3b, 0xa3, 0xed, 0xfd, 0x7a, 0x7b, 0x12, 0xb2,
			0x7a, 0xc7, 0x2c, 0x3e, 0x67, 0x76, 0x8f, 0x61,
			0x7f, 0xc8, 0x1b, 0xc3, 0x88, 0x8a, 0x51, 0x32,
			0x3a, 0x9f, 0xb8, 0xaa, 0x4b, 0x1e, 0x5e, 0x4a,
		},
	)
	// genesisBlock defines
	// genesisBlock defines the genesis block of the block chain which serves as the
	// public transaction ledger for the main network.
	genesisBlock = wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:    1,
			PrevBlock:  chainhash.Hash{},  // 0000000000000000000000000000000000000000000000000000000000000000
			MerkleRoot: genesisMerkleRoot, // 4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b
			Timestamp: time.Unix(
				0x495fab29,
				0,
			), // 2009-01-03 18:15:05 +0000 UTC
			Bits:  0x1d00ffff, // 486604799 [00000000ffff0000000000000000000000000000000000000000000000000000]
			Nonce: 0x7c2bac1d, // 2083236893
		},
		Transactions: []*wire.MsgTx{&genesisCoinbaseTx},
	}
)
