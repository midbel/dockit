package oxml

import (
	"encoding/xml"
)

const wbBaseDir = "xl"

const (
	typeSheetUrl  = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet"
	typeDocUrl    = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument"
	typeMainUrl   = "http://schemas.openxmlformats.org/spreadsheetml/2006/main"
	typeSharedUrl = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/sharedStrings"
)

const (
	mimeRels         = "application/vnd.openxmlformats-package.relationships+xml"
	mimeXml          = "application/xml"
	mimeWorkbook     = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"
	mimeWorksheet    = "application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"
	mimeStyle        = "application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"
	mimeSharedString = "application/vnd.openxmlformats-officedocument.spreadsheetml.sharedStrings+xml"
)

type xmlWorkbook struct{}

type xmlSheet struct {
	XMLName xml.Name   `xml:"sheet"`
	Id      string     `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
	Name    string     `xml:"name,attr"`
	Index   int        `xml:"sheetId,attr"`
	State   SheetState `xml:"state,attr"`
}

type xmlRelations struct {
	XMLName   xml.Name      `xml:"Relationships"`
	Xmlns     string        `xml:"xmlns,attr"`
	Relations []xmlRelation `xml:"Relationship"`
}

type xmlRelation struct {
	XMLName xml.Name `xml:"Relationship"`
	Target  string   `xml:",attr"`
	Id      string   `xml:",attr"`
	Type    string   `xml:",attr"`
}

type xmlSharedStrings struct {
	XMLName   xml.Name          `xml:"sst"`
	Xmlns     string            `xml:"xmlns,attr"`
	Count     int               `xml:"count,attr"`
	UniqCount int               `xml:"uniqueCount,attr"`
	Values    []xmlSharedString `xml:"si"`
}

type xmlSharedString struct {
	Value string `xml:"t"`
}

type xmlRow struct {
	XMLName xml.Name `xml:"row"`
	Line    int64    `xml:"r,attr"`
	Cells   []*Cell  `xml:"c"`
}

type xmlFormula struct {
	XMLName xml.Name `xml:"f"`
	Type    string   `xml:"t,attr"`
	Index   string   `xml:"si,attr"`
	Ref     string   `xml:"ref,attr"`
	Expr    string   `xml:",chardata"`
}
