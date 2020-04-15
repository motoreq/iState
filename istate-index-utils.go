//

package istate

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"math"
	"strconv"
	"strings"
)

type compactIndexV map[string]string

func removeKeysWithStar(keyValue map[string][]byte) {
	for index := range keyValue {
		if strings.Contains(index, star) {
			delete(keyValue, index)
		}
	}
}

func removeStarFromKeys(keyValue map[string][]byte) {
	for index := range keyValue {
		// Replace is used as ReplaceAll isnt available in go version used in fabric image
		newIndex := strings.Replace(index, star, "", len(index))
		if newIndex != index {
			keyValue[newIndex] = keyValue[index]
			delete(keyValue, index)
		}

	}
}

func removeStarFromKey(key string) (newKey string) {
	// Replace is used as ReplaceAll isnt available in go version used in fabric image
	newKey = strings.Replace(key, star, "", len(key))
	return
}

func addKeyWithoutOverLap(query []map[string]interface{}, index string, value interface{}) (newQuery []map[string]interface{}) {
	newQuery = query
	successFlag := false
	for i := 0; i < len(newQuery); i++ {
		if _, ok := newQuery[i][index]; !ok {
			newQuery[i][index] = value
			successFlag = true
			break
		}
	}
	if !successFlag {
		tempMap := make(map[string]interface{})
		tempMap[index] = value
		newQuery = append(newQuery, tempMap)
	}
	return
}

func fetchCompactIndex(stub shim.ChaincodeStubInterface, key string) (val compactIndexV, iStateErr Error) {
	valBytes, err := stub.GetState(key)
	if err != nil {
		iStateErr = NewError(err, 5001)
		return
	}
	if valBytes == nil {
		return
	}
	err = json.Unmarshal(valBytes, &val)
	if err != nil {
		iStateErr = NewError(err, 5002)
		return
	}
	return
}

func putCompactIndex(stub shim.ChaincodeStubInterface, cIndex map[string]compactIndexV) (iStateErr Error) {

	for index, val := range cIndex {
		mv, err := json.Marshal(val)
		if err != nil {
			iStateErr = NewError(err, 5003)
			return
		}
		err = stub.PutState(index, mv)
		if err != nil {
			iStateErr = NewError(err, 5004)
			return
		}
	}
	return
}

func generateCIndexKey(index string) (compactIndex string, keyRef string) {
	compactIndex, keyRef = splitIndexAndKey(index)
	compactIndex = removeLastSeparator(compactIndex)
	return
}

func deriveIndexKeys(indexKey string, isQuery bool) (derivedKeys []string) {
	splitParts := strings.Split(indexKey, seperator)
	if len(splitParts) < 4 {
		return
	}
	middleParts := splitParts[2 : len(splitParts)-1]
	prefix := strings.Join(splitParts[:2], seperator)
	suffix := splitParts[len(splitParts)-1]
	derivedKeys = deriveIndexPermutation(middleParts, prefix, suffix, isQuery)

	return
}
func deriveIndexPermutation(vals []string, prefix string, suffix string, isQuery bool) (permuteds []string) {
	numDigits := len(vals)
	maxCount := int(math.Pow(2, float64(numDigits)))
	// We don't want 1111 -> which is already main index
	permuteds = make([]string, maxCount-2)
	for i := 1; i < maxCount-1; i++ {
		permString := fmt.Sprintf("%v", strconv.FormatInt(int64(i), 2))

		// Fill zeros
		diff := numDigits - len(permString)
		if diff > 0 {
			bs := make([]byte, diff)
			for i := 0; i < diff; i++ {
				bs[i] = '0'
			}
			permString = string(bs) + permString
		}
		newIndex := asciiLast + removeSuffixZeros(permString) + seperator + prefix + seperator + getIndexPermVal(vals, permString, isQuery) + seperator + suffix
		permuteds[i-1] = newIndex
	}

	return
}

func removeSuffixZeros(val string) (removed string) {
	removed = val
	for i := len(val) - 1; i > -1; i-- {
		if val[i] != '0' {
			if i+1 < len(val) {
				removed = removed[:i+1]
			}
			break
		}
	}
	return
}

func getIndexPermVal(vals []string, permString string, isQuery bool) (permVal string) {
	permVal = ""
	for i := 0; i < len(permString); i++ {
		presetFlag := false
		if permString[i] == '1' {
			permVal += vals[i]
			presetFlag = true
		}
		switch isQuery && !presetFlag {
		case true:
			permVal += star + seperator
		default:
			permVal += seperator
		}
	}
	// Remove last seperator
	permVal = permVal[:len(permVal)-len(seperator)]
	return
}

func removeNValsFromIndex(index string, n int) (partIndex string, removedVals []string) {
	partIndex = index
	removedVals = make([]string, n, n)
	seperatorLen := len(seperator)
	for i := 0; i < n; i++ {
		lastIndex := strings.LastIndex(partIndex, seperator)
		if lastIndex == -1 {
			return
		}
		switch lastIndex+seperatorLen >= len(partIndex) {
		case true:
			removedVals[i] = ""
		default:
			removedVals[i] = partIndex[lastIndex+seperatorLen:] // separator + null == 2 chars
		}
		partIndex = partIndex[:lastIndex]
	}
	partIndex = partIndex + seperator
	return
}
