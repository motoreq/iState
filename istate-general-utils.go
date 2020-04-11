//

package istate

import (
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"reflect"
	"strings"
)

//
func (iState *iState) setStub(stub *shim.ChaincodeStubInterface) {
	iState.currentStub = stub
	return
}

func (iState *iState) unmarshalToStruct(valBytes []byte) (uObj reflect.Value, iStateErr Error) {
	singleElem := reflect.New(reflect.TypeOf(iState.structRef)).Interface()
	err := json.Unmarshal(valBytes, &singleElem)
	if err != nil {
		iStateErr = NewError(err, 4004)
		return
	}
	uObj = reflect.ValueOf(singleElem).Elem()
	return
}

func (iState *iState) getQIndexMap(key string, valBytes []byte) (encodedKV map[string][]byte, iStateErr Error) {
	var tempVar map[string]interface{}
	err := json.Unmarshal(valBytes, &tempVar)
	if err != nil {
		iStateErr = NewError(err, 4005)
		return
	}
	encodedKV, _, _, iStateErr = iState.encodeState(tempVar, key, "", 1) // keyRefSeperatedIndex = 1, query = false
	if iStateErr != nil {
		return
	}
	return
}

//
func convertObjToMap(obj interface{}) (uObj map[string]interface{}, iStateErr Error) {
	mo, err := json.Marshal(obj)
	if err != nil {
		iStateErr = NewError(err, 4001)
		return
	}
	err = json.Unmarshal(mo, &uObj)
	if err != nil {
		iStateErr = NewError(err, 4002)
		return
	}
	return
}

//
func getKeyByRange(stub shim.ChaincodeStubInterface, startKey, endKey string, limit ...int) (fetchedKVMap map[string][]byte, iStateErr Error) {
	if len(limit) == 0 {
		limit = []int{int32Biggest}
	}
	fetchedKVMap = make(map[string][]byte)
	iterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		iStateErr = NewError(err, 3006)
		return
	}
	defer iterator.Close()
	for i := 0; iterator.HasNext(); i++ {
		iteratorResult, err := iterator.Next()
		if err != nil {
			iStateErr = NewError(err, 4003)
			return
		}
		key := iteratorResult.GetKey()
		val := iteratorResult.GetValue()
		fetchedKVMap[key] = val

		if i >= limit[0] {
			break
		}
	}
	return
}

func getKeyByRangeWithPagination(stub shim.ChaincodeStubInterface, startKey, endKey string, pagesize int32, bookmark string) (fetchedKVMap map[string][]byte, iStateErr Error) {
	fetchedKVMap = make(map[string][]byte)
	iterator, _, err := stub.GetStateByRangeWithPagination(startKey, endKey, pagesize, bookmark)
	if err != nil {
		iStateErr = NewError(err, 3006)
		return
	}
	defer iterator.Close()
	for i := 0; iterator.HasNext(); i++ {
		iteratorResult, err := iterator.Next()
		if err != nil {
			iStateErr = NewError(err, 4003)
			return
		}
		key := iteratorResult.GetKey()
		val := iteratorResult.GetValue()
		fetchedKVMap[key] = val
	}
	return
}
func getKeyFromIndex(indexkey string) (keyRef string) {
	splitPosition := strings.LastIndex(indexkey, null)
	if splitPosition != -1 {
		keyRef = indexkey[splitPosition+1:]
	}
	return
}

func splitIndexAndKey(index string) (partindex, keyRef string) {
	partindex = index
	splitPosition := strings.LastIndex(index, null)
	if splitPosition != -1 {
		partindex = index[:splitPosition]
		keyRef = index[splitPosition+1:]
	}
	return
}
