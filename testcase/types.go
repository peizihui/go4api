/*
 * go4api - a api testing tool written in Go
 * Created by: Ping Zhu 2018
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 *
 */

package testcase

import (
)

// test case data type, includes testcase
type TestCaseDataInfo struct {
    TestCase *TestCase
    JsonFilePath string
    CsvFile string
    CsvRow string
    MutationInfo interface{}
}

// test case execution type, includes testdata
type TestCaseExecutionInfo struct {
    TestCaseDataInfo *TestCaseDataInfo
    TestResult string  // Ready, Running, Success, Fail, ParentReady, ParentRunning, ParentFailed
    ActualStatusCode int
    StartTime string
    EndTime string
    TestMessages string
    StartTimeUnixNano int64
    EndTimeUnixNano int64
    DurationUnixNano int64
}
type TestCases []TestCase

// test case type,
type TestCase map[string]*TestCaseBasics

type TestCaseBasics struct {
    Priority string         `json:"priority"`
    ParentTestCase string   `json:"parentTestCase"`
    Inputs []interface{}     `json:"inputs"`
    Request *Request         `json:"request"`
    Response *Response       `json:"response"`
    Outputs []interface{}   `json:"outputs"`
}

type Request struct {  
    Method string                       `json:"method"`
    Path string                         `json:"path"`
    Headers map[string]interface{}      `json:"headers"`
    QueryString map[string]interface{}  `json:"queryString"`
    Payload map[string]interface{}      `json:"payload"`
}


type Response struct {  
    Status map[string]interface{}   `json:"status"`
    Headers map[string]interface{}  `json:"headers"`
    Body map[string]interface{}     `json:"body"`
}

// for report format 
type TcReportResults struct { 
    TcName string 
    Priority string
    ParentTestCase string
    JsonFilePath string
    CsvFile string
    CsvRow string
    MutationInfo interface{}
    TestResult string  // Ready, Running, Success, Fail, ParentReady, ParentRunning, ParentFailed
    ActualStatusCode int
    StartTime string
    EndTime string
    TestMessages string
    StartTimeUnixNano int64
    EndTimeUnixNano int64
    DurationUnixNano int64
}


type TcConsoleResults struct { 
    TcName string 
    Priority string
    ParentTestCase string
    JsonFilePath string
    CsvFile string
    CsvRow string
    MutationInfo interface{}
    TestResult string  // Ready, Running, Success, Fail, ParentReady, ParentRunning, ParentFailed
    ActualStatusCode int
    TestMessages string
}

