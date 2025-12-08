package oxml

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type writer struct {
	base   string
	writer *zip.Writer
	io.Closer
}

func writeFile(file string) (*writer, error) {
	w, err := os.Create(file)
	if err != nil {
		return nil, err
	}
	z := writer{
		base:   "xl",
		writer: zip.NewWriter(w),
		Closer: w,
	}
	return &z, nil
}

func (w *writer) WriteFile(file *File) error {
	for _, s := range file.sheets {
		s.addr = w.createTarget("worksheets", s.Name+".xml")
		if err := w.writeWorksheet(s); err != nil {
			return err
		}
	}
	if err := w.writeWorkbook(file); err != nil {
		return err
	}
	if err := w.writeRelationForSheets(file); err != nil {
		return err
	}
	if err := w.writeRelations(); err != nil {
		return err
	}
	if err := w.writeSharedStrings(); err != nil {
		return err
	}
	if err := w.writeStyles(); err != nil {
		return err
	}
	if err := w.writeContentTypes(file); err != nil {
		return err
	}
	return nil
}

func (w *writer) Close() error {
	w.writer.Close()
	return w.Closer.Close()
}

func (z *writer) writeContentTypes(file *File) error {
	w, err := z.writer.Create("[Content_Types].xml")
	if err != nil {
		return err
	}

	type xmlDefault struct {
		XMLName     xml.Name `xml:"Default"`
		Extension   string   `xml:"Extension,attr"`
		ContentType string   `xml:"ContentType,attr"`
	}

	type xmlOverride struct {
		XMLName     xml.Name `xml:"Override"`
		PartName    string   `xml:"PartName,attr"`
		ContentType string   `xml:"ContentType,attr"`
	}

	root := struct {
		XMLName   xml.Name      `xml:"Types"`
		Xmlns     string        `xml:"xmlns,attr"`
		Defaults  []xmlDefault  `xml:"Default"`
		Overrides []xmlOverride `xml:"Override"`
	}{
		Xmlns: "http://schemas.openxmlformats.org/package/2006/content-types",
		Defaults: []xmlDefault{
			{
				Extension:   "rels",
				ContentType: mimeRels,
			},
			{
				Extension:   "xml",
				ContentType: mimeXml,
			},
		},
		Overrides: []xmlOverride{
			{
				PartName:    "/xl/workbook.xml",
				ContentType: mimeWorkbook,
			},
			{
				PartName:    "/xl/sharedStrings.xml",
				ContentType: mimeSharedString,
			},
			{
				PartName:    "/xl/styles.xml",
				ContentType: mimeStyle,
			},
		},
	}
	for _, s := range file.sheets {
		ox := xmlOverride{
			PartName:    "/" + s.addr,
			ContentType: mimeWorksheet,
		}
		root.Overrides = append(root.Overrides, ox)
	}
	return xml.NewEncoder(w).Encode(&root)
}

func (z *writer) writeStyles() error {
	w, err := z.writer.Create(z.createTarget("styles.xml"))
	if err != nil {
		return err
	}
	root := struct {
		XMLName xml.Name `xml:"styleSheet"`
		Xmlns   string   `xml:"xmlns,attr"`
	}{
		Xmlns: typeMainUrl,
	}
	return xml.NewEncoder(w).Encode(&root)
}

func (z *writer) writeSharedStrings() error {
	w, err := z.writer.Create(z.createTarget("sharedStrings.xml"))
	if err != nil {
		return err
	}

	root := struct {
		XMLName     xml.Name `xml:"sst"`
		Xmlns       string   `xml:"xmlns,attr"`
		Count       int      `xml:"count"`
		uniqueCount int      `xml:"uniqueCount,attr"`
	}{
		Xmlns: typeMainUrl,
	}

	return xml.NewEncoder(w).Encode(&root)
}

func (z *writer) writeRelations() error {
	w, err := z.writer.Create("_rels/.rels")
	if err != nil {
		return err
	}

	type xmlRelation struct {
		Id     string `xml:",attr"`
		Type   string `xml:",attr"`
		Target string `xml:",attr"`
	}

	root := struct {
		XMLName   xml.Name      `xml:"Relationships"`
		Xmlns     string        `xml:"xmlns,attr"`
		Relations []xmlRelation `xml:"Relationship"`
	}{
		Xmlns: "http://schemas.openxmlformats.org/package/2006/relationships",
	}

	root.Relations = append(root.Relations, xmlRelation{
		Id:     "rId1",
		Type:   typeDocUrl,
		Target: "xl/workbook.xml",
	})
	return xml.NewEncoder(w).Encode(&root)
}

func (z *writer) writeRelationForSheets(f *File) error {
	w, err := z.writer.Create(z.createTarget("_rels", "workbook.xml.rels"))
	if err != nil {
		return err
	}

	type xmlRelation struct {
		Id     string `xml:"Id,attr"`
		Type   string `xml:",attr"`
		Target string `xml:",attr"`
	}

	root := struct {
		XMLName   xml.Name      `xml:"Relationships"`
		Xmlns     string        `xml:"xmlns,attr"`
		Relations []xmlRelation `xml:"Relationship"`
	}{
		Xmlns: "http://schemas.openxmlformats.org/package/2006/relationships",
	}
	for _, s := range f.sheets {
		target, err := filepath.Rel("xl", s.addr)
		if err != nil {
			return err
		}
		rx := xmlRelation{
			Id:     s.Id,
			Type:   typeSheetUrl,
			Target: strings.ReplaceAll(target, "\\", "/"),
		}
		root.Relations = append(root.Relations, rx)
	}
	return xml.NewEncoder(w).Encode(&root)
}

func (z *writer) writeWorksheet(sheet *Sheet) error {
	w, err := z.writer.Create(sheet.addr)
	if err != nil {
		return err
	}

	type xmlRow struct {
		XMLName xml.Name `xml:"row"`
		Line    int64    `xml:"r,attr"`
		Cells   []*Cell  `xml:"c"`
	}

	root := struct {
		XMLName   xml.Name `xml:"worksheet"`
		Xmlns     string   `xml:"xmlns,attr"`
		RelXmlns  string   `xml:"xmlns:r,attr"`
		Dimension struct {
			Ref string `xml:"ref,attr"`
		} `xml:"dimension"`
		Rows []xmlRow `xml:"sheetData>row"`
	}{
		Xmlns:    typeMainUrl,
		RelXmlns: "http://schemas.openxmlformats.org/officeDocument/2006/relationships",
	}
	start, end := sheet.Bounding()
	root.Dimension.Ref = fmt.Sprintf("%s:%s", start.Addr(), end.Addr())
	for _, r := range sheet.Rows {
		rx := xmlRow{
			Line:  r.Line,
			Cells: r.Cells,
		}
		root.Rows = append(root.Rows, rx)
	}
	return xml.NewEncoder(w).Encode(&root)
}

func (z *writer) writeWorkbook(f *File) error {
	w, err := z.writer.Create(z.createTarget("workbook.xml"))
	if err != nil {
		return err
	}
	type xmlSheet struct {
		XMLName xml.Name `xml:"sheet"`
		Id      string   `xml:"r:id,attr"`
		Name    string   `xml:"name,attr"`
		Index   int      `xml:"sheetId,attr"`
	}
	root := struct {
		XMLName  xml.Name   `xml:"workbook"`
		Xmlns    string     `xml:"xmlns,attr"`
		RelXmlns string     `xml:"xmlns:r,attr"`
		Sheets   []xmlSheet `xml:"sheets>sheet"`
	}{
		Xmlns:    typeMainUrl,
		RelXmlns: "http://schemas.openxmlformats.org/officeDocument/2006/relationships",
	}
	for _, s := range f.sheets {
		xs := xmlSheet{
			Id:    s.Id,
			Index: s.Index,
			Name:  s.Name,
		}
		root.Sheets = append(root.Sheets, xs)
	}
	return xml.NewEncoder(w).Encode(&root)
}

func (w *writer) createTarget(parts ...string) string {
	parts = append([]string{w.base}, parts...)
	return strings.Join(parts, "/")
}
