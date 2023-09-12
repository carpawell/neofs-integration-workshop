package workshop

import (
	"github.com/nspcc-dev/neo-go/pkg/interop"
	"github.com/nspcc-dev/neo-go/pkg/interop/contract"
	"github.com/nspcc-dev/neo-go/pkg/interop/native/management"
	"github.com/nspcc-dev/neo-go/pkg/interop/native/oracle"
	"github.com/nspcc-dev/neo-go/pkg/interop/native/std"
	"github.com/nspcc-dev/neo-go/pkg/interop/runtime"
	"github.com/nspcc-dev/neo-go/pkg/interop/storage"
)

const ownerKey = "owner"
const objectKey = "object"

func _deploy(data interface{}, _ bool) {
	owner := data.(interop.Hash160)
	storage.Put(storage.GetContext(), ownerKey, owner)
}

func GetObject() []byte {
	return storage.Get(storage.GetReadOnlyContext(), objectKey).([]byte)
}

func SaveObject(cid, oid string) {
	checkAccess()

	url := "neofs:" + cid + "/" + oid
	oracle.Request(url, nil, "saveObjectCB", nil, 15*oracle.MinimumResponseGas)
}

func SaveObjectCB(url string, data interface{}, code int, result []byte) {
	if string(runtime.GetCallingScriptHash()) != oracle.Hash {
		panic("called from non-oracle contract")
	}
	if code != oracle.Success {
		panic("request failed for " + url + " with code " + std.Itoa(code, 10))
	}

	storage.Put(storage.GetContext(), objectKey, result)
}

func RemoveObject() {
	checkAccess()

	storage.Delete(storage.GetContext(), objectKey)
}

func Owner() interop.Hash160 {
	return storage.Get(storage.GetReadOnlyContext(), ownerKey).(interop.Hash160)
}

func Update(nef []byte, manifest string, data interface{}) {
	checkAccess()

	contract.Call(interop.Hash160(management.Hash), "update",
		contract.All, nef, manifest, data)
}

func checkAccess() {
	owner := storage.Get(storage.GetReadOnlyContext(), ownerKey).(interop.Hash160)
	if !runtime.CheckWitness(owner) {
		panic("not allowed user")
	}
}
