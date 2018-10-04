package main

import (
	"github.com/hornbill/goApiLib"
)

const (
	version = "1.0.0"
)

var (
	apiCallConfig  apiCallStruct
	boolConfLoaded bool
	configFileName string
	davEndpoint    string
	espXmlmc       *apiLib.XmlmcInstStruct
	logFile        string
	timeNow        string
)

type apiCallStruct struct {
	APIKey     string
	InstanceID string
	Reports    []reportStruct
}

type reportStruct struct {
	ReportID             int
	ReportName           string
	DeleteReportInstance bool
	ReportFolder         string
}

type stateStruct struct {
	Code     string `xml:"code"`
	ErrorRet string `xml:"error"`
}

type xmlmcReportResponse struct {
	MethodResult string      `xml:"status,attr"`
	State        stateStruct `xml:"state"`
	RunID        int         `xml:"params>runId"`
}

type xmlmcReportStatusResponse struct {
	MethodResult string             `xml:"status,attr"`
	State        stateStruct        `xml:"state"`
	Params       paramsReportStruct `xml:"params"`
}

type paramsReportStruct struct {
	ReportRun reportRunStruct    `xml:"reportRun"`
	Files     []reportFileStruct `xml:"files"`
}

type reportRunStruct struct {
	RunID    int    `xml:"runId"`
	ReportID int    `xml:"reportId"`
	Status   string `xml:"status"`
	RunBy    string `xml:"runBy"`
	CSVLink  string `xml:"csvLink"`
}

type reportFileStruct struct {
	Name string `xml:"name"`
	Type string `xml:"type"`
}
