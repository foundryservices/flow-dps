package wal

import (
	"fmt"

	"github.com/onflow/flow-go/ledger"
	"github.com/onflow/flow-go/ledger/common/encoding"
	"github.com/onflow/flow-go/ledger/common/utils"
)

type Operation uint8

const (
	OperationUpdate Operation = 1
	OperationDelete Operation = 2
)

/*
The LedgerWAL update record uses two operations so far - an update which must include all keys and values, and deletion
which only needs a root tree state commitment.
Updates need to be atomic, hence we prepare binary representation of whole change set.
Since keys, values and state commitments date types are variable length, we have to store it as well.
Every record has:

1 byte Operation Type | 2 bytes Big Endian uint16 length of state commitment | state commitment data

If OP = OperationUpdate, then it follows with:

4 bytes Big Endian uint32 - total number of key/value pairs | 2 bytes Big Endian uint16 - length of key (keys are the same length)

and for every pair after
bytes for key | 4 bytes Big Endian uint32 - length of value | value bytes

The code here is deliberately simple, for performance.

*/

func Decode(data []byte) (operation Operation, rootHash ledger.RootHash, update *ledger.TrieUpdate, err error) {
	if len(data) < 4 { // 1 byte op + 2 size + actual data = 4 minimum
		err = fmt.Errorf("data corrupted, too short to represent operation - hexencoded data: %x", data)
		return
	}

	operation = Operation(data[0])
	switch operation {
	case OperationUpdate:
		update, err = encoding.DecodeTrieUpdate(data[1:])
		return
	case OperationDelete:
		var rootHashBytes []byte
		rootHashBytes, _, err = utils.ReadShortData(data[1:])
		if err != nil {
			err = fmt.Errorf("cannot read state commitment: %w", err)
			return
		}
		rootHash, err = ledger.ToRootHash(rootHashBytes)
		if err != nil {
			err = fmt.Errorf("invalid root hash: %w", err)
			return
		}
		return
	default:
		err = fmt.Errorf("unknown operation type, given: %x", operation)
		return
	}
}
