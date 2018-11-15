/*
 * go4api - a api testing tool written in Go
 * Created by: Ping Zhu 2018
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 *
 */

package api

import (
    "fmt"
    "strings" 
    "encoding/json"

    "go4api/lib/testcase"

    gjson "github.com/tidwall/gjson"
)

func (tcDataStore *TcDataStore) CommandGroup (cmdGroupOrigin []*testcase.CommandDetails) (string, [][]*testcase.TestMessage) {
    finalResults := "Success"
    var cmdsResults []bool
    var finalTestMessages = [][]*testcase.TestMessage{}
    //
    // cmdGroupJsonOriginB, _ := json.Marshal(cmdGroupOrigin)
    // cmdGroupJsonOrigin := string(cmdGroupJsonOriginB)

    for i := 0; i < tcDataStore.CmdGroupLength; i ++ {
        cmdType := cmdGroupOrigin[i].CmdType

        switch strings.ToLower(cmdType) {
            case "sql":
                cmdGroupJson := tcDataStore.PrepCmd(i, ".cmd")
                //
                cmdStr := gjson.Get(cmdGroupJson, fmt.Sprint(i) + "." + "cmd")
                // init
                tcDataStore.CmdType = "sql"
                tcDataStore.CmdExecStatus = ""
                tcDataStore.CmdAffectedCount = -1
                tcDataStore.CmdResults = -1

                cmdAffectedCount, _, cmdResults, cmdExecStatus := RunSql(cmdStr.String())
                fmt.Println(">>>>: ", cmdStr, cmdAffectedCount, cmdResults, cmdExecStatus)
                tcDataStore.CmdExecStatus = cmdExecStatus
                tcDataStore.CmdAffectedCount = cmdAffectedCount
                tcDataStore.CmdResults = cmdResults

                cmdsResults, finalTestMessages = tcDataStore.HandleSingleCmdResult(i)
            case "redis":
                var cmdStr, cmdKey, cmdValue string

                cmdGroupJson := tcDataStore.PrepCmd(i, ".cmd")
                //
                cmdMap := gjson.Get(cmdGroupJson, fmt.Sprint(i) + "." + "cmd").Map()
       
                for k, v := range cmdMap {
                    cmdStr = k
                    if len(v.Array()) == 1 {
                        cmdKey = v.Array()[0].String()
                        cmdValue = ""
                    }
                    if len(v.Array()) > 1 {
                        cmdKey = v.Array()[0].String()
                        cmdValue = v.Array()[1].String()
                    }
                }
                // init
                tcDataStore.CmdType = "redis"
                tcDataStore.CmdExecStatus = ""
                tcDataStore.CmdAffectedCount = -1
                tcDataStore.CmdResults = -1

                cmdAffectedCount, cmdResults, cmdExecStatus := RunRedis(cmdStr, cmdKey, cmdValue)
                
                tcDataStore.CmdExecStatus = cmdExecStatus
                tcDataStore.CmdAffectedCount = cmdAffectedCount
                tcDataStore.CmdResults = cmdResults

                cmdsResults, finalTestMessages = tcDataStore.HandleSingleCmdResult(i)
            default:
                fmt.Println("!! warning, command ", cmdType, " can not be recognized.")
        }
    }

    for key := range cmdsResults {
        if cmdsResults[key] == false {
            finalResults = "Fail"
            break
        }
    }

    return finalResults, finalTestMessages
}

func (tcDataStore *TcDataStore) PrepCmd (i int, subPath string) string {
    var cmdGroup []*testcase.CommandDetails

    cmdGroupJsonB, _ := json.Marshal(tcDataStore.TcData)
    cmdGroupJson := string(cmdGroupJsonB)

    cmdStrPath := "TestCase." + tcDataStore.TcData.TestCase.TcName() + "." + tcDataStore.CmdSection + "." + fmt.Sprint(i) + subPath
    tcDataStore.RenderTcVariables(cmdStrPath)

    cmdGroupJsonB, _ = json.Marshal(tcDataStore.TcData)
    cmdGroupJson = string(cmdGroupJsonB)

    tcDataStore.EvaluateTcBuiltinFunctions(cmdStrPath)

    cmdGroupJsonB, _ = json.Marshal(tcDataStore.TcData)
    cmdGroupJson = string(cmdGroupJsonB)

    switch tcDataStore.CmdSection {
        case "setUp":
            cmdGroup = tcDataStore.TcData.TestCase.SetUp()
        case "tearDown":
            cmdGroup = tcDataStore.TcData.TestCase.TearDown()
    }

    cmdGroupJsonB, _ = json.Marshal(cmdGroup)
    cmdGroupJson = string(cmdGroupJsonB)

    return cmdGroupJson
}

func (tcDataStore *TcDataStore) HandleSingleCmdResult (i int) ([]bool, [][]*testcase.TestMessage) {
    // --------
    var cmdsResults []bool
    var finalTestMessages = [][]*testcase.TestMessage{}

    if tcDataStore.CmdExecStatus == "cmdSuccess" {
        cmdGroupJson := tcDataStore.PrepCmd(i, ".cmdResponse")
        //
        cmdExpResp := gjson.Get(cmdGroupJson, fmt.Sprint(i) + "." + "cmdResponse").Map()

        singleCmdResults, testMessages := tcDataStore.CompareRespGroup(cmdExpResp)

        // HandleSingleCommandResults for out
        if singleCmdResults == true {
            tcDataStore.HandleCmdResultsForOut(i)
        }

        cmdsResults = append(cmdsResults, singleCmdResults)
        finalTestMessages = append(finalTestMessages, testMessages)
    } else {
        cmdsResults = append(cmdsResults, false)
    }

    return cmdsResults, finalTestMessages
}

func (tcDataStore *TcDataStore) CompareRespGroup (cmdExpResp map[string]gjson.Result) (bool, []*testcase.TestMessage){
    //-----------
    singleCmdResults := true
    var testResults []bool
    var testMessages []*testcase.TestMessage

    for key, value := range cmdExpResp {
        cmdExpResp_sub := value.Value().(map[string]interface{})
        for assertionKey, expValueOrigin := range cmdExpResp_sub {
            
            actualValue := tcDataStore.GetResponseValue(key)

            var expValue interface{}
            switch expValueOrigin.(type) {
                case float64, int64: 
                    expValue = expValueOrigin
                default:
                    expValue = tcDataStore.GetResponseValue(expValueOrigin.(string))
            }
            
            testRes, msg := compareCommon(tcDataStore.CmdType, key, assertionKey, actualValue, expValue)
            
            testMessages = append(testMessages, msg)
            testResults = append(testResults, testRes)
        }
    }

    for key := range testResults {
        if testResults[key] == false {
            singleCmdResults = false
            break
        }
    }

    return singleCmdResults, testMessages
}

func (tcDataStore *TcDataStore) HandleCmdResultsForOut (i int) {
    var cmdGroup []*testcase.CommandDetails

    // write out session if has
    cmdGroup = tcDataStore.PrepCmdGroup(i, ".session")
    expTcSession := cmdGroup[i].Session
    tcDataStore.WriteSession(expTcSession)

    // write out global variables if has
    cmdGroup = tcDataStore.PrepCmdGroup(i, ".outGlobalVariables")
    expOutGlobalVariables := cmdGroup[i].OutGlobalVariables
    tcDataStore.WriteOutGlobalVariables(expOutGlobalVariables)

    // write out tc loca variables if has
    cmdGroup = tcDataStore.PrepCmdGroup(i, ".outLocalVariables")
    expOutLocalVariables := cmdGroup[i].OutLocalVariables
    tcDataStore.WriteOutGlobalVariables(expOutLocalVariables)
}

func (tcDataStore *TcDataStore) PrepCmdGroup (i int, subPath string) []*testcase.CommandDetails {
    var cmdGroup []*testcase.CommandDetails

    tcDataStore.PrepCmd(i, subPath)

    switch tcDataStore.CmdSection {
        case "setUp":
            cmdGroup = tcDataStore.TcData.TestCase.SetUp()
        case "tearDown":
            cmdGroup = tcDataStore.TcData.TestCase.TearDown()
    }

    return cmdGroup
}