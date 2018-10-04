package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/hornbill/goApiLib"
	"github.com/hornbill/goHornbillHelpers"
)

func main() {
	timeNow = time.Now().Format("2006-01-02 15:04:05")
	logFile = "reporting_" + time.Now().Format("20060102150405") + ".log"
	flag.StringVar(&configFileName, "file", "conf.json", "Name of the configuration file to load")
	flag.Parse()

	hornbillHelpers.Logger(3, "---- Hornbill Report Runner V"+fmt.Sprintf("%v", version)+" ----", true, logFile)
	hornbillHelpers.Logger(3, "Flag - Configuration File: "+fmt.Sprintf("%s", configFileName), true, logFile)

	//-- Load Configuration File Into Struct
	apiCallConfig, boolConfLoaded = loadConfig()
	if boolConfLoaded != true {
		hornbillHelpers.Logger(4, "Unable to load config, process closing.", true, logFile)
		return
	}
	//Global XMLMC session
	espXmlmc = apiLib.NewXmlmcInstance(apiCallConfig.InstanceID)
	espXmlmc.SetAPIKey(apiCallConfig.APIKey)

	davEndpoint = apiLib.GetEndPointFromName(apiCallConfig.InstanceID) + "/dav/"

	//Run and get report content
	for _, definition := range apiCallConfig.Reports {
		runReport(definition, espXmlmc)
	}

}

func runReport(report reportStruct, espXmlmc *apiLib.XmlmcInstStruct) {

	hornbillHelpers.Logger(3, "Running report "+report.ReportName+" ["+strconv.Itoa(report.ReportID)+"].", true, logFile)

	espXmlmc.SetParam("reportId", strconv.Itoa(report.ReportID))
	espXmlmc.SetParam("comment", "Run from the goHornbillReport tool")

	XMLMC, xmlmcErr := espXmlmc.Invoke("reporting", "reportRun")
	if xmlmcErr != nil {
		hornbillHelpers.Logger(4, xmlmcErr.Error(), true, logFile)
		return
	}

	var xmlRespon xmlmcReportResponse

	err := xml.Unmarshal([]byte(XMLMC), &xmlRespon)
	if err != nil {
		hornbillHelpers.Logger(4, fmt.Sprintf("%v", err), true, logFile)
		return
	}
	if xmlRespon.MethodResult != "ok" {
		hornbillHelpers.Logger(4, xmlRespon.State.ErrorRet, true, logFile)
		return
	}
	if xmlRespon.RunID > 0 {
		reportComplete := false
		for reportComplete == false {
			reportSuccess, reportComplete, reportDetails := checkReport(xmlRespon.RunID, espXmlmc)

			if reportComplete == true {
				if reportSuccess == false {
					return
				}
				getReportContent(reportDetails, espXmlmc, report)
				if report.DeleteReportInstance {
					deleteReportInstance(reportDetails.ReportRun.RunID)
				}
				return
			}
			time.Sleep(time.Second * 10)
		}
	} else {
		hornbillHelpers.Logger(4, "No RunID returned from ", true, logFile)
		return
	}
}

func checkReport(runID int, espXmlmc *apiLib.XmlmcInstStruct) (bool, bool, paramsReportStruct) {

	hornbillHelpers.Logger(3, "Checking report run for completion.", true, logFile)
	espXmlmc.SetParam("runId", strconv.Itoa(runID))
	XMLMC, xmlmcErr := espXmlmc.Invoke("reporting", "reportRunGetStatus")

	if xmlmcErr != nil {
		hornbillHelpers.Logger(4, xmlmcErr.Error(), true, logFile)
		return false, true, paramsReportStruct{}
	}

	var xmlRespon xmlmcReportStatusResponse

	err := xml.Unmarshal([]byte(XMLMC), &xmlRespon)
	if err != nil {
		hornbillHelpers.Logger(4, fmt.Sprintf("%v", err), true, logFile)
		return false, true, paramsReportStruct{}
	}
	if xmlRespon.MethodResult != "ok" {
		hornbillHelpers.Logger(4, xmlRespon.State.ErrorRet, true, logFile)
		return false, true, paramsReportStruct{}
	}

	switch xmlRespon.Params.ReportRun.Status {
	case "pending":
		fallthrough
	case "started":
		fallthrough
	case "running":
		return false, false, paramsReportStruct{}
	case "completed":
		return true, true, xmlRespon.Params
	case "failed":
		fallthrough
	case "aborted":
		return false, true, paramsReportStruct{}
	}
	return false, false, paramsReportStruct{}
}

func getReportContent(reportOutput paramsReportStruct, espXmlmc *apiLib.XmlmcInstStruct, report reportStruct) {
	for _, v := range reportOutput.Files {
		getFile(reportOutput.ReportRun, v, espXmlmc, report)
	}
}

func getFile(reportRun reportRunStruct, file reportFileStruct, espXmlmc *apiLib.XmlmcInstStruct, report reportStruct) {
	hornbillHelpers.Logger(3, "Retrieving "+file.Type+" Report File "+file.Name, true, logFile)
	//Create file for data dump
	out, err := os.Create(report.ReportFolder + "/" + file.Name)
	if err != nil {
		hornbillHelpers.Logger(4, fmt.Sprintf("%v", err), true, logFile)
	}
	defer out.Close()
	reportURL := davEndpoint + "reports/" + strconv.Itoa(reportRun.ReportID) + "/" + reportRun.CSVLink

	req, _ := http.NewRequest("GET", reportURL, nil)
	req.Header.Set("Content-Type", "text/xmlmc")
	req.Header.Set("Authorization", "ESP-APIKEY "+apiCallConfig.APIKey)

	if err != nil {
		hornbillHelpers.Logger(4, fmt.Sprintf("%v", err), true, logFile)
	}
	duration := time.Second * time.Duration(30)
	client := &http.Client{Timeout: duration}

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	//-- Check for HTTP Response
	if resp.StatusCode != 200 {
		hornbillHelpers.Logger(4, fmt.Sprintf("Invalid HTTP Response: %d", resp.StatusCode), true, logFile)
		io.Copy(ioutil.Discard, resp.Body)
		return
	}
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		hornbillHelpers.Logger(4, fmt.Sprintf("%v", err), true, logFile)
		return
	}
	hornbillHelpers.Logger(3, "[SUCCESS] Retrieved report data ["+report.ReportFolder+"/"+file.Name+"]", true, logFile)
	return
}
func deleteReportInstance(runID int) {
	hornbillHelpers.Logger(3, "Deleting report run instance.", true, logFile)
	espXmlmc.SetParam("runId", strconv.Itoa(runID))
	XMLMC, xmlmcErr := espXmlmc.Invoke("reporting", "reportRunDelete")

	if xmlmcErr != nil {
		hornbillHelpers.Logger(4, xmlmcErr.Error(), true, logFile)
		return
	}

	var xmlRespon xmlmcReportStatusResponse

	err := xml.Unmarshal([]byte(XMLMC), &xmlRespon)
	if err != nil {
		hornbillHelpers.Logger(4, fmt.Sprintf("%v", err), true, logFile)
		return
	}
	if xmlRespon.MethodResult != "ok" {
		hornbillHelpers.Logger(4, xmlRespon.State.ErrorRet, true, logFile)
		return
	}
	hornbillHelpers.Logger(3, "[SUCCESS] Report run instance deleted.", true, logFile)
}

//loadConfig -- Function to Load Configruation File
func loadConfig() (apiCallStruct, bool) {
	boolLoadConf := true
	//-- Check Config File File Exists
	cwd, _ := os.Getwd()
	configurationFilePath := cwd + "/" + configFileName
	hornbillHelpers.Logger(1, "Loading Config File: "+configurationFilePath, false, logFile)
	if _, fileCheckErr := os.Stat(configurationFilePath); os.IsNotExist(fileCheckErr) {
		hornbillHelpers.Logger(4, "No Configuration File", true, logFile)
		os.Exit(102)
	}
	//-- Load Config File
	file, fileError := os.Open(configurationFilePath)
	//-- Check For Error Reading File
	if fileError != nil {
		hornbillHelpers.Logger(4, "Error Opening Configuration File: "+fmt.Sprintf("%v", fileError), true, logFile)
		boolLoadConf = false
	}

	//-- New Decoder
	decoder := json.NewDecoder(file)
	//-- New Var based on apiCallStruct
	edbConf := apiCallStruct{}
	//-- Decode JSON
	err := decoder.Decode(&edbConf)
	//-- Error Checking
	if err != nil {
		hornbillHelpers.Logger(4, "Error Decoding Configuration File: "+fmt.Sprintf("%v", err), true, logFile)
		boolLoadConf = false
	}
	//-- Return New Config
	return edbConf, boolLoadConf
}
