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
	writer *zip.Writer
	io.Closer
}

func writeFile(file string) (*writer, error) {
	w, err := os.Create(file)
	if err != nil {
		return nil, err
	}
	z := writer{
		writer: zip.NewWriter(w),
		Closer: w,
	}
	return &z, nil
}

func (w *writer) WriteFile(file *File) error {
	for _, s := range file.sheets {
		target := []string{
			"xl",
			"worksheets",
			s.Name + ".xml",
		}
		s.addr = strings.Join(target, "/")
		if err := writeWorksheet(w.writer, s); err != nil {
			return err
		}
	}
	if err := writeWorkbook(w.writer, file); err != nil {
		return err
	}
	if err := writeRelationForSheets(w.writer, file); err != nil {
		return err
	}
	if err := writeRelations(w.writer); err != nil {
		return err
	}
	if err := writeSharedStrings(w.writer); err != nil {
		return err
	}
	if err := writeStyles(w.writer); err != nil {
		return err
	}
	if err := writeContentTypes(w.writer, file); err != nil {
		return err
	}
	return nil
}

func (w *writer) Close() error {
	w.writer.Close()
	return w.Closer.Close()
}

func writeContentTypes(z *zip.Writer, f *File) error {
	w, err := z.Create("[Content_Types].xml")
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
	for _, s := range f.sheets {
		ox := xmlOverride{
			PartName:    "/" + s.addr,
			ContentType: mimeWorksheet,
		}
		root.Overrides = append(root.Overrides, ox)
	}
	return xml.NewEncoder(w).Encode(&root)
}

func writeStyles(z *zip.Writer) error {
	target := []string{
		"xl",
		"styles.xml",
	}
	w, err := z.Create(strings.Join(target, "/"))
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

func writeSharedStrings(z *zip.Writer) error {
	target := []string{
		"xl",
		"sharedStrings.xml",
	}
	w, err := z.Create(strings.Join(target, "/"))
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

func writeRelations(z *zip.Writer) error {
	target := []string{
		"_rels",
		".rels",
	}
	w, err := z.Create(strings.Join(target, "/"))
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

func writeRelationForSheets(z *zip.Writer, f *File) error {
	target := []string{
		"xl",
		"_rels",
		"workbook.xml.rels",
	}
	w, err := z.Create(strings.Join(target, "/"))
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

func writeWorksheet(z *zip.Writer, sheet *Sheet) error {
	w, err := z.Create(sheet.addr)
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

func writeWorkbook(z *zip.Writer, f *File) error {
	target := []string{
		"xl",
		"workbook.xml",
	}
	w, err := z.Create(strings.Join(target, "/"))
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
