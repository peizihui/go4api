/*
 * go4api - a api testing tool written in Go
 * Created by: Ping Zhu 2018
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 *
 */
 
package texttmpl

import (
    // "fmt"
    // "io/ioutil"                                                                                                                                              
    "os"
    // "strings"
    "bytes"
    "text/template"
    // "time"

    "go4api/utils"
)

type ResultsJs struct {
    GStart_time int64
    GStart   string
    PEnd_time int64
    PEnd  string
    TcReportStr string
}

// var setUp_gStartUnixNano = 1539337198430831188
// var setUp_gStart = "2018-10-12 17:39:58.430831188 +0800 CST m=+0.002470518"
// var setUp_pEndUnixNano = 1539337200123399284
// var setUp_pEnd = "2018-10-12 17:40:00.123399284 +0800 CST m=+1.695102131"
// var setUp_tcResults = ExecutionResultSlice

// var gStartUnixNano = 1539337198430831188
// var gStart = "2018-10-12 17:39:58.430831188 +0800 CST m=+0.002470518"
// var pEndUnixNano = 1539337200123399284
// var pEnd = "2018-10-12 17:40:00.123399284 +0800 CST m=+1.695102131"
// var tcResults = ExecutionResultSlice

// var tearDown_gStartUnixNano = 1539337198430831188
// var tearDown_gStart = "2018-10-12 17:39:58.430831188 +0800 CST m=+0.002470518"
// var tearDown_pEndUnixNano = 1539337200123399284
// var tearDown_pEnd = "2018-10-12 17:40:00.123399284 +0800 CST m=+1.695102131"
// var tearDown_tcResults = ExecutionResultSlice

type ResultsJss struct {
    setUp_gStartUnixNano int64
    setUp_gStart   string
    setUp_pEndUnixNano int64
    setUp_pEnd  string
    setUp_tcResults string

    gStartUnixNano int64
    gStart   string
    pEndUnixNano int64
    pEnd  string
    tcResults string

    tearDown_gStartUnixNano int64
    tearDown_gStart   string
    tearDown_pEndUnixNano int64
    tearDown_pEnd  string
    tearDown_tcResults string
}

type DetailsJs struct {
    StatsStr string
}

type StatsJs struct {
    StatsStr_1 string
    StatsStr_2 string
    StatsStr_Success string
    StatsStr_Fail string
}

type MStatsJs struct {
    StatsStr_1 string
    StatsStr_2 string
    StatsStr_3 string
}


func GetTemplateFromString() {
    type Inventory struct {
        Material string
        Count    uint
    }
    sweaters := Inventory{"wool", 17}
    tmpl := template.Must(template.New("test").Parse("{{.Count}} of {{.Material}} \n"))

    err := tmpl.Execute(os.Stdout, sweaters)
    if err != nil {
      panic(err) 
    }
}


func GenerateDetailsJs(strVar string, targetFile string, detailsJs *DetailsJs, logResultsFile string) {
    outFile, err := os.OpenFile(targetFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
       panic(err) 
    }
    defer outFile.Close()
    //
    tmpl := template.Must(template.New("HtmlJsCss").Parse(strVar))

    err = tmpl.Execute(outFile, *detailsJs)
    if err != nil {
      panic(err) 
    }
}

func GenerateResultsJs(strVar string, targetFile string, resultsJs *ResultsJs, logResultsFile string) {
    outFile, err := os.OpenFile(targetFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
       panic(err) 
    }
    defer outFile.Close()
    //
    tmpl := template.Must(template.New("HtmlJsCss").Parse(strVar))

    err = tmpl.Execute(outFile, *resultsJs)
    if err != nil {
      panic(err) 
    }
}


func GenerateStatsJs(strVar string, targetFile string, resultsJs []string, logResultsFile string) {
    outFile, err := os.OpenFile(targetFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
       panic(err) 
    }
    defer outFile.Close()
    //
    tmpl := template.Must(template.New("HtmlJsCss").Parse(strVar))

    statsJs := StatsJs {
        StatsStr_1: resultsJs[0],
        StatsStr_2: resultsJs[1],
        StatsStr_Success: resultsJs[2],
        StatsStr_Fail: resultsJs[3],
    }

    err = tmpl.Execute(outFile, statsJs)
    if err != nil {
      panic(err) 
    }
}


func GenerateMutationResultsJs(strVar string, targetFile string, resultsJs []string, logResultsFile string) {
    outFile, err := os.OpenFile(targetFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
       panic(err) 
    }
    defer outFile.Close()
    //
    tmpl := template.Must(template.New("HtmlJsCss").Parse(strVar))

    mStatsJs := MStatsJs {
        StatsStr_1: resultsJs[0],
        StatsStr_2: resultsJs[1],
        StatsStr_3: resultsJs[2],
    }

    err = tmpl.Execute(outFile, mStatsJs)
    if err != nil {
      panic(err) 
    }
}

func GenerateJsonBasedOnTemplateAndCsv(jsonFilePath string, testData map[string]interface{}) *bytes.Buffer {
    jsonTemplateBytes := utils.GetContentFromFile(jsonFilePath)
    //
    tcJson := GetTcJson(string(jsonTemplateBytes), testData)

    return tcJson
}

func GetTcJson (jsonTemplate string, testData map[string]interface{}) *bytes.Buffer {
    tmpl := template.Must(template.New("tcTemp").Parse(jsonTemplate))
    
    tcJson := &bytes.Buffer{}
    // Execute the template
    err := tmpl.Execute(tcJson, testData)
    if err != nil {
      panic(err) 
    }

    return tcJson
}


