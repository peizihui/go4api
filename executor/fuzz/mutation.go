/*
 * go4api - a api testing tool written in Go
 * Created by: Ping Zhu 2018
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 *
 */

package fuzz

import (                                                                                                                                             
    // "os"
    // "time"
    "fmt"
    // "path/filepath"
    "strings"
    // "strconv"
    "reflect"
    "encoding/json"
    "go4api/testcase"
    // gjson "github.com/tidwall/gjson"
    sjson "github.com/tidwall/sjson"
)

// mutation is to mutate the valid data to working api, see if mutated invalid data still can be handled by the api
// two ways to mutate the testcase:
// Option 1: 
// copy the underlying fields and values to another TestCaseDataInfo, with mutation(s)
// the better way would be deep copy the TestCaseDataInfo, and change the value, but Golang standard
// Lib has no deepcopy, so that, here use a plain way, that is, to re-sturct the TestCaseDataInfo
//
// Option 2:
// json.Marshal the tc in originMutationTcArray, 
// then change the value(s) in the json
// then Unmarshal the to testcase, and add to mutatedTcArray
// then execute the mutatedTcArray

// focus on the Request to mutate
// type Request struct {  
//     Method string
//     Path string
//     Headers map[string]interface{}
//     QueryString map[string]interface{}
//     Payload map[string]interface{}
// }

type PayloadInfo struct {
    FieldPath []string
    CurrValue interface{}
    FieldType string // the json supported types
    FieldSubType string  // like ip/email/phone/etc.
}

func MutateTcArray(originMutationTcArray []testcase.TestCaseDataInfo) []testcase.TestCaseDataInfo {
    var mutatedTcArray []testcase.TestCaseDataInfo

    for _, originTcData := range originMutationTcArray {
        tcJson, _ := json.Marshal(originTcData)
        // fmt.Println("tcJson:", string(tcJson)) 

        mutatedTcArray = append(mutatedTcArray, originTcData)

        // here to start the mutation
        // headers
        mutatedTcArray = append(mutatedTcArray, MutateSetRequestHeader(tcJson))
        mutatedTcArray = append(mutatedTcArray, MutateAddRequestHeader(tcJson))

        i := 0
        for k, _ := range originTcData.TestCase.ReqHeaders() {
            mutatedTcArray = append(mutatedTcArray, MutateDelRequestHeader(tcJson, k, i))
            i = i + 1
        }

        // querystring
        i = 0
        for key, value := range originTcData.TestCase.ReqQueryString() {
            fmt.Println(reflect.TypeOf(value))
            // if value match number mode
            mutatedTcArray = append(mutatedTcArray, MutateSetRequestQueryString(tcJson, key, fmt.Sprint(-1), key + fmt.Sprint(i)))
            i = i + 1
            mutatedTcArray = append(mutatedTcArray, MutateSetRequestQueryString(tcJson, key, fmt.Sprint(0), key + fmt.Sprint(i)))
            i = i + 1
            mutatedTcArray = append(mutatedTcArray, MutateSetRequestQueryString(tcJson, key, fmt.Sprint(10000), key + fmt.Sprint(i)))
            i = i + 1
        }


        mutatedTcArray = append(mutatedTcArray, MutateAddRequestQueryString(tcJson))

        i = 0
        for key, _ := range originTcData.TestCase.ReqQueryString() {
            mutatedTcArray = append(mutatedTcArray, MutateDelRequestQueryString(tcJson, key, i))
            i = i + 1
        }


        // --------------------------------------------
        // Payload, strategy, loop all node, mutate it (include remove)
        //
        // payloadJson, _ := json.Marshal(originTcData.TestCase.ReqPayload())
        // fmt.Println("->payloadJson: ", string(payloadJson)) 

        for key, value := range originTcData.TestCase.ReqPayload() {
            if key == "text" {
                

                // fmt.Println(" ---> to loop over the struct and display")
                c := make(chan PayloadInfo)

                go func(c chan PayloadInfo) {
                    defer close(c)
                    sturctFieldsMutation(c, []string{}, value)
                }(c)

                i := 0
                for keyPath := range c {
                    // fmt.Println("keyPath: ", keyPath)

                    // get the value
                    // result := gjson.Get(string(payloadJson), key + "." + strings.Join(keyPath.FieldPath, "."))

                    // set the value
                    payloadPath := key + "." + strings.Join(keyPath.FieldPath, ".")
                    payloadFullPath := "TestCase." + originTcData.TcName() + ".Request.Payload" + "." + payloadPath

                    // mutate the value based on rules
                    // get values 
                    mutatedValues := MutateFunc(keyPath.FieldPath, keyPath.CurrValue, keyPath.FieldType, keyPath.FieldSubType)
                    // remove node

                    for _, mutatedValue := range mutatedValues {
                        i = i + 1
                        mutatedTcJson, _ := sjson.Set(string(tcJson), payloadFullPath, mutatedValue)
                        mutationInfo := fmt.Sprint(keyPath) + "," + fmt.Sprint(keyPath.CurrValue) + ", `" + fmt.Sprint(mutatedValue) + "`"
                        mutatedTcArray = append(mutatedTcArray, MutatePayload([]byte(mutatedTcJson), mutationInfo, "1-" + fmt.Sprint(i)))
                    }

                    // (1) -> empty ("")
                    // (2) -> blank (" ")
                    // (3) -> prefix blank (" " + previousValue)
                    // (4) -> suffix blank (previousValue + " ")
                    // (5) -> mid blank (previousValue[0:2] + " " + previousValue[2:])
                    // (6) -> only one char (previousValue[0])
                    // (7) -> longlong string (strings.Repeat(previousValue, 50)
                    // (8) -> special char(s) (~!@#$%^&*()_+{}[]<>?)
                    // (9) -> null
                    // (10) -> change type (i.e. to int/float64/...)
                }
            }
        }
    }
    // fmt.Println("\nmutatedTcArray: ", mutatedTcArray
    return mutatedTcArray
}


func MutateFunc(fieldPath []string, currValue interface{}, fieldType string, fieldSubType string) []interface{} {
    var mutatedValues []interface{}

    // if fieldType && fieldSubType

    mutatedValue := ""
    mutatedValues = append(mutatedValues, mutatedValue)

    mutatedValue = " "
    mutatedValues = append(mutatedValues, mutatedValue)

    mutatedValue = " " + fmt.Sprint(currValue)
    mutatedValues = append(mutatedValues, mutatedValue)

    mutatedValue = fmt.Sprint(currValue) + " "
    mutatedValues = append(mutatedValues, mutatedValue)

    mutatedValue = fmt.Sprint(currValue)[0:2] + " " + fmt.Sprint(currValue)[2:]
    mutatedValues = append(mutatedValues, mutatedValue)

    return mutatedValues
}


func MutateSetRequestHeader (tcJson []byte) testcase.TestCaseDataInfo {
    var mTcData testcase.TestCaseDataInfo
    json.Unmarshal(tcJson, &mTcData)

    originTcName := mTcData.TcName()
    mTcData.TestCase = mTcData.TestCase.UpdateTcName(originTcName + "-M-seth-" + fmt.Sprint(1))
    mTcData.TestCase.SetRequestHeader("aaaa", "dbddsdsfa")

    return mTcData
}


func MutateAddRequestHeader (tcJson []byte) testcase.TestCaseDataInfo {
    var mTcData testcase.TestCaseDataInfo
    json.Unmarshal(tcJson, &mTcData)

    originTcName := mTcData.TcName()
    mTcData.TestCase = mTcData.TestCase.UpdateTcName(originTcName + "-M-addh-" + fmt.Sprint(2))
    mTcData.TestCase.AddRequestHeader("aaaakk", "dbddsdsfa")

    return mTcData
}


func MutateDelRequestHeader (tcJson []byte, k string, i int) testcase.TestCaseDataInfo {
    var mTcData testcase.TestCaseDataInfo
    json.Unmarshal(tcJson, &mTcData)

    originTcName := mTcData.TcName()
    mTcData.TestCase = mTcData.TestCase.UpdateTcName(originTcName + "-M-Delh-" + fmt.Sprint(i))
    mTcData.TestCase.DelRequestHeader(k)

    return mTcData
}



//
func MutateSetRequestQueryString (tcJson []byte, key string, value string, suffix string) testcase.TestCaseDataInfo {
    var mTcData testcase.TestCaseDataInfo
    json.Unmarshal(tcJson, &mTcData)

    originTcName := mTcData.TcName()
    mTcData.TestCase = mTcData.TestCase.UpdateTcName(originTcName + "-M-setq-" + suffix)
    mTcData.TestCase.SetRequestQueryString(key, value)

    return mTcData
}


func MutateAddRequestQueryString (tcJson []byte) testcase.TestCaseDataInfo {
    var mTcData testcase.TestCaseDataInfo
    json.Unmarshal(tcJson, &mTcData)

    originTcName := mTcData.TcName()
    mTcData.TestCase = mTcData.TestCase.UpdateTcName(originTcName + "-M-addq-" + fmt.Sprint(2))

    mTcData.TestCase.AddRequestQueryString("aaaakk", "dbddsdsfa")

    return mTcData
}


func MutateDelRequestQueryString (tcJson []byte, k string, i int) testcase.TestCaseDataInfo {
    var mTcData testcase.TestCaseDataInfo
    json.Unmarshal(tcJson, &mTcData)

    originTcName := mTcData.TcName()
    mTcData.TestCase = mTcData.TestCase.UpdateTcName(originTcName + "-M-Delq-" + fmt.Sprint(i))

    mTcData.TestCase.DelRequestQueryString(k)

    return mTcData
}


//
func MutatePayload (tcJson []byte, mutationInfo interface{}, suffix string) testcase.TestCaseDataInfo {
    var mTcData testcase.TestCaseDataInfo
    json.Unmarshal(tcJson, &mTcData)

    // change the tc name
    originTcName := mTcData.TcName()
    mTcData.TestCase = mTcData.TestCase.UpdateTcName(originTcName + "-M-P-" + suffix)
    // change the tc priority?

    mTcData.MutationInfo = mutationInfo

    return mTcData
}


func sturctFieldsDisplay(value interface{}) {
    switch reflect.TypeOf(value).Kind() {
        case reflect.String:
            fmt.Println("value: ", value, reflect.TypeOf(value), reflect.TypeOf(value).Kind())
        case reflect.Int32:
            fmt.Println("value: ", value, reflect.TypeOf(value), reflect.TypeOf(value).Kind())
        case reflect.Map: {
            // fmt.Println("value: ", value, reflect.TypeOf(value), reflect.TypeOf(value).Kind())
            for key2, value2 := range reflect.ValueOf(value).Interface().(map[string]interface{}) {
                // fmt.Println("key2, value2: ", key2, reflect.TypeOf(value2))
                switch reflect.TypeOf(value2).Kind() {
                    case reflect.String:
                        fmt.Println("key2, value2: ", key2, value2, reflect.TypeOf(value2), reflect.TypeOf(value2).Kind())
                    case reflect.Int32:
                        fmt.Println("key2, value2: ", key2, value2, reflect.TypeOf(value2), reflect.TypeOf(value2).Kind())
                    case reflect.Map:
                        fmt.Println("key2, value2: ", key2, value2, reflect.TypeOf(value2), reflect.TypeOf(value2).Kind())
                        sturctFieldsDisplay(value2)
                    case reflect.Array:
                        // note: maybe the Array/Slice is the last node, if it contains concrete type, like [1, 2, 3, ...]
                        fmt.Println("key2, value2: ", key2, value2, reflect.TypeOf(value2), reflect.TypeOf(value2).Kind())
                        sturctFieldsDisplay(value2)
                    case reflect.Slice:
                        // note: maybe the Array/Slice is the last node, if it contains concrete type, like [1, 2, 3, ...]
                        fmt.Println("key2, value2: ", key2, value2, reflect.TypeOf(value2), reflect.TypeOf(value2).Kind())
                        sturctFieldsDisplay(value2)
                }
            }     
        }
        case reflect.Array, reflect.Slice: {
            // fmt.Println("value: ", value, reflect.TypeOf(value), reflect.TypeOf(value).Kind())
            for key2, value2 := range reflect.ValueOf(value).Interface().([]interface{}) {
                // fmt.Println("key2, value2: ", key2, reflect.TypeOf(value2))
                switch reflect.TypeOf(value2).Kind() {
                    case reflect.String:
                        fmt.Println("key2, value2: ", key2, value2, reflect.TypeOf(value2), reflect.TypeOf(value2).Kind())
                    case reflect.Int32:
                        fmt.Println("key2, value2: ", key2, value2, reflect.TypeOf(value2), reflect.TypeOf(value2).Kind())
                    case reflect.Map:
                        fmt.Println("key2, value2: ", key2, value2, reflect.TypeOf(value2), reflect.TypeOf(value2).Kind())
                        sturctFieldsDisplay(value2)
                    case reflect.Array:
                        // note: maybe the Array/Slice is the last node, if it contains concrete type, like [1, 2, 3, ...]
                        fmt.Println("key2, value2: ", key2, value2, reflect.TypeOf(value2), reflect.TypeOf(value2).Kind())
                        sturctFieldsDisplay(value2)
                    case reflect.Slice:
                        // note: maybe the Array/Slice is the last node, if it contains concrete type, like [1, 2, 3, ...]
                        fmt.Println("key2, value2: ", key2, value2, reflect.TypeOf(value2), reflect.TypeOf(value2).Kind())
                        sturctFieldsDisplay(value2)
                }
            }  
        }
    }
}


func sturctFieldsMutation(c chan PayloadInfo, subPath []string, value interface{}) {
    switch reflect.TypeOf(value).Kind() {
        case reflect.Map: {
            // fmt.Println("value: ", value, reflect.TypeOf(value), reflect.TypeOf(value).Kind())
            for key2, value2 := range reflect.ValueOf(value).Interface().(map[string]interface{}) {
                // fmt.Println("key2, value2: ", key2, reflect.TypeOf(value2))
                switch reflect.TypeOf(value2).Kind() {
                    case reflect.String, reflect.Int, reflect.Float64:
                        subPathNew := append(subPath, key2)
                        output := make([]string, len(subPathNew))
                        copy(output, subPathNew)

                        payloadPath := PayloadInfo{output, value2, reflect.TypeOf(value2).Kind().String(), ""}
                        c <- payloadPath
                    case reflect.Map:
                        subPathNew := append(subPath, key2)
                        sturctFieldsMutation(c, subPathNew, value2)
                    case reflect.Array, reflect.Slice:
                        // note: maybe the Array/Slice is the last node, if it contains concrete type, like [1, 2, 3, ...]
                        for _, v := range reflect.ValueOf(value2).Interface().([]interface{}) {
                            switch reflect.TypeOf(v).Kind() {
                                case reflect.Array, reflect.Slice, reflect.Map:
                                    subPathNew := append(subPath, fmt.Sprint(key2))
                                    sturctFieldsMutation(c, subPathNew, value2)
                            }
                            break
                        }
                }
            }     
        }
        case reflect.Array, reflect.Slice: {
            // fmt.Println("value: ", value, reflect.TypeOf(value), reflect.TypeOf(value).Kind())
            for key2, value2 := range reflect.ValueOf(value).Interface().([]interface{}) {
                // fmt.Println("key2, value2: ", key2, reflect.TypeOf(value2))
                switch reflect.TypeOf(value2).Kind() {
                    case reflect.String, reflect.Int, reflect.Float64:
                        subPathNew := append(subPath, fmt.Sprint(key2))
                        output := make([]string, len(subPathNew))
                        copy(output, subPathNew)

                        payloadPath := PayloadInfo{output, value2, reflect.TypeOf(value2).Kind().String(), ""}
                        c <- payloadPath
                    case reflect.Map:
                        subPathNew := append(subPath, fmt.Sprint(key2))
                        sturctFieldsMutation(c, subPathNew, value2)
                    case reflect.Array, reflect.Slice:
                        // note: maybe the Array/Slice is the last node, if it contains concrete type, like [1, 2, 3, ...]
                        for _, v := range reflect.ValueOf(value2).Interface().([]interface{}) {
                            switch reflect.TypeOf(v).Kind() {
                                case reflect.Array, reflect.Slice, reflect.Map:
                                    subPathNew := append(subPath, fmt.Sprint(key2))
                                    sturctFieldsMutation(c, subPathNew, value2)
                            }
                        }
                        break
                }
            } 
        }
    }
}

