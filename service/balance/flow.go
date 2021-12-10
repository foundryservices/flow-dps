package balance

import (
	"encoding/hex"
	"fmt"
	"math"
	"strings"

	"github.com/fxamacker/cbor/v2"
	"github.com/onflow/atree"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/interpreter"
	executionState "github.com/onflow/flow-go/engine/execution/state"
	"github.com/onflow/flow-go/fvm/state"
	"github.com/onflow/flow-go/ledger"
	"github.com/onflow/flow-go/model/flow"
)

type FlowFinder struct {
	vaultTypeID string
	storage     atree.SlabStorage
}

func (ff *FlowFinder) DetectFlow(path ledger.Path, p *ledger.Payload, flows map[flow.Address]map[ledger.Path]uint64) error {

	//id, err := keyToRegisterID(p.Key)
	//if err != nil {
	//	return err
	//}
	//
	//// Ignore known payload keys that are not Cadence values
	//if state.IsFVMStateKey(id.Owner, id.Controller, id.Key) {
	//	return nil
	//}
	//
	//value, version := interpreter.StripMagic(p.Value)
	//
	//err = storageMigrationV5DecMode.Valid(value)
	//if err != nil {
	//	return err
	//}
	//
	//decodeFunction := interpreter.DecodeValue
	//if version <= 4 {
	//	decodeFunction = interpreter.DecodeValueV4
	//}
	//
	//// Decode the value
	//owner := common.BytesToAddress([]byte(id.Owner))
	//cPath := []string{id.Key}
	//
	//cValue, err := decodeFunction(value, &owner, cPath, version, nil)
	//if err != nil {
	//	return fmt.Errorf(
	//		"failed to decode value: %w\n\nvalue:\n%s\n",
	//		err, hex.Dump(value),
	//	)
	//}
	//
	////if id.Key == "contract\u001fFlowToken" {
	////	tokenSupply := uint64(cValue.(*interpreter.CompositeValue).GetField("totalSupply").(interpreter.UFix64Value))
	////	r.Log.Info().Uint64("tokenSupply", tokenSupply).Msg("total token supply")
	////	r.totalSupply = tokenSupply
	////}
	//
	////flows := make(map[flow.Address]map[string]uint64)
	//
	//balanceVisitor := &interpreter.EmptyVisitor{
	//	CompositeValueVisitor: func(inter *interpreter.Interpreter, value *interpreter.CompositeValue) bool {
	//
	//		if string(value.TypeID()) == "A.1654653399040a61.FlowToken.Vault" {
	//			b := uint64(value.GetField("balance").(interpreter.UFix64Value))
	//			address := flow.BytesToAddress([]byte(id.Owner))
	//
	//			if helpers.IsDebugAccount(address, helpers.DebugFind) {
	//
	//				fmt.Printf("Found %d flow for %s under path %x => %x/%x/%s\n", b, address.String(), path[:], id.Owner, id.Controller, id.Key)
	//				//fmt.Printf("current balances for address: \n")
	//				//for path, b := range flows[address] {
	//				//	fmt.Printf("%x => %d\n", path[:], b)
	//				//}
	//			}
	//
	//			if _, has := flows[address]; !has {
	//				flows[address] = make(map[ledger.Path]uint64)
	//			}
	//
	//			flows[address][path] += b
	//
	//			return false
	//		}
	//		return true
	//	},
	//	DictionaryValueVisitor: func(interpreter *interpreter.Interpreter, value *interpreter.DictionaryValue) bool {
	//		return value.DeferredKeys() == nil
	//	},
	//}
	//
	//inter, err := interpreter.NewInterpreter(nil, common.StringLocation("somewhere"))
	//if err != nil {
	//	return err
	//}
	//cValue.Accept(inter, balanceVisitor)
	//
	//return nil

	/////////////////////////////

	id, err := keyToRegisterID(p.Key)
	if err != nil {
		return err
	}

	// Ignore known payload keys that are not Cadence values
	if state.IsFVMStateKey(id.Owner, id.Controller, id.Key) {
		return nil
	}
	if !(strings.HasPrefix(id.Key, common.PathDomainPublic.Identifier()) &&
		strings.HasPrefix(id.Key, common.PathDomainPrivate.Identifier()) &&
		strings.HasPrefix(id.Key, common.PathDomainStorage.Identifier())) {
		// this is not a storage path
		return nil
	}

	//owner := common.BytesToAddress([]byte(id.Owner))
	decoder := interpreter.CBORDecMode.NewByteStreamDecoder(p.Value)

	storable, err := interpreter.DecodeStorable(decoder, atree.StorageIDUndefined)
	if err != nil || storable == nil {

		return fmt.Errorf(
			"failed to decode storable: %w\n\nvalue:\n%s\n",
			err, hex.Dump(p.Value),
		)
	}
	storedValue, err := storable.StoredValue(ff.storage)
	cValue := interpreter.MustConvertStoredValue(storedValue)
	if err != nil || cValue == nil {

		return fmt.Errorf(
			"failed to decode value: %w\n\nvalue:\n%s\n",
			err, hex.Dump(p.Value),
		)
	}

	//if id.Key == "contract\u001fFlowToken" {
	//	tokenSupply := uint64(cValue.(*interpreter.CompositeValue).GetField(nil, nil, "totalSupply").(interpreter.UFix64Value))
	//	r.Log.Info().Uint64("tokenSupply", tokenSupply).Msg("total token supply")
	//}

	//lastComposite := "none"
	//firstComposite := ""

	balanceVisitor := &interpreter.EmptyVisitor{
		CompositeValueVisitor: func(inter *interpreter.Interpreter, value *interpreter.CompositeValue) bool {

			if string(value.TypeID()) == ff.vaultTypeID {
				b := uint64(value.GetField(nil, nil, "balance").(interpreter.UFix64Value))
				address := flow.BytesToAddress([]byte(id.Owner))
				if b == 0 {
					// ignore 0 balance results
					return false
				}

				if _, has := flows[address]; !has {
					flows[address] = make(map[ledger.Path]uint64)
				}

				flows[address][path] += b

				return false
			}

			return true

			//////////////
		},
	}

	inter, err := interpreter.NewInterpreter(nil, common.StringLocation("somewhere"))
	if err != nil {

		return fmt.Errorf("could not create interpreter: %w for %x", err, []byte(id.Key))
	}
	cValue.Accept(inter, balanceVisitor)

	return nil
}

func keyToRegisterID(key ledger.Key) (flow.RegisterID, error) {
	if len(key.KeyParts) != 3 ||
		key.KeyParts[0].Type != executionState.KeyPartOwner ||
		key.KeyParts[1].Type != executionState.KeyPartController ||
		key.KeyParts[2].Type != executionState.KeyPartKey {
		return flow.RegisterID{}, fmt.Errorf("key not in expected format %s", key.String())
	}

	return flow.NewRegisterID(
		string(key.KeyParts[0].Value),
		string(key.KeyParts[1].Value),
		string(key.KeyParts[2].Value),
	), nil
}

var storageMigrationV5DecMode = func() cbor.DecMode {
	decMode, err := cbor.DecOptions{
		IntDec:           cbor.IntDecConvertNone,
		MaxArrayElements: math.MaxInt32,
		MaxMapPairs:      math.MaxInt32,
		MaxNestedLevels:  256,
	}.DecMode()
	if err != nil {
		panic(err)
	}
	return decMode
}()
