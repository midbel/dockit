package oxml

import (
	"archive/zip"
	"compress/flate"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	sax "github.com/midbel/codecs/xml"
)

const startIx = 1000

type writer struct {
	base   string
	writer *zip.Writer
	io.Closer

	lastUsedId int
	err        error
}

func writeFile(file string) (*writer, error) {
	w, err := os.Create(file)
	if err != nil {
		return nil, err
	}
	z := writer{
		base:       wbBaseDir,
		writer:     zip.NewWriter(w),
		Closer:     w,
		lastUsedId: startIx,
	}
	z.writer.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestCompression)
	})
	return &z, nil
}

func (z *writer) WriteFile(file *File) error {
	for _, s := range file.sheets {
		z.writeWorksheet(s)
		if z.invalid() {
			return z.err
		}
	}
	z.writeWorkbook(file)
	z.writeSharedStrings(file)
	z.writeRelationForSheets(file)
	z.writeRelations()
	z.writeStyles()
	z.writeContentTypes(file)
	return z.err
}

func (w *writer) Close() error {
	w.writer.Close()
	return w.Closer.Close()
}

func (z *writer) writeContentTypes(file *File) {
	if z.invalid() {
		return
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
		addr := z.createTarget("worksheets", fmt.Sprintf("%s.xml", s.Name))
		ox := xmlOverride{
			PartName:    "/" + z.fromBase(addr),
			ContentType: mimeWorksheet,
		}
		root.Overrides = append(root.Overrides, ox)
	}
	z.encodeXML("[Content_Types].xml", &root)
}

func (z *writer) writeStyles() {
	if z.invalid() {
		return
	}
	root := struct {
		XMLName xml.Name `xml:"styleSheet"`
		Xmlns   string   `xml:"xmlns,attr"`
	}{
		Xmlns: typeMainUrl,
	}
	z.encodeXML("styles.xml", root)
}

func (z *writer) writeSharedStrings(file *File) {
	if len(file.sharedStrings) == 0 {
		return
	}
	if z.invalid() {
		return
	}
	root := xmlSharedStrings{
		Xmlns:     typeMainUrl,
		Count:     len(file.sharedStrings),
		UniqCount: len(file.sharedStrings),
	}
	for _, s := range file.sharedStrings {
		shared := xmlSharedString{
			Value: s,
		}
		root.Values = append(root.Values, shared)
	}
	z.encodeXML("xl/sharedStrings.xml", &root)
}

func (z *writer) writeRelations() {
	if z.invalid() {
		return
	}
	root := xmlRelations{
		Xmlns: "http://schemas.openxmlformats.org/package/2006/relationships",
		Relations: []xmlRelation{
			{
				Id:     "rId1",
				Type:   typeDocUrl,
				Target: "xl/workbook.xml",
			},
		},
	}
	z.encodeXML("_rels/.rels", &root)
}

func (z *writer) writeRelationForSheets(file *File) {
	if z.invalid() {
		return
	}

	root := xmlRelations{
		Xmlns: "http://schemas.openxmlformats.org/package/2006/relationships",
	}
	for _, sh := range file.sheets {
		addr := z.createTarget("worksheets", fmt.Sprintf("%s.xml", sh.Name))
		rx := xmlRelation{
			Id:     sh.Id,
			Type:   typeSheetUrl,
			Target: z.fromBase(addr),
		}
		root.Relations = append(root.Relations, rx)
	}
	if len(file.sharedStrings) > 0 {
		rx := xmlRelation{
			Id:     z.createFileID(),
			Type:   typeSharedUrl,
			Target: "sharedStrings.xml",
		}
		root.Relations = append(root.Relations, rx)
	}
	addr := z.createTarget("_rels", "workbook.xml.rels")
	z.encodeXML(addr, &root)
}

func (z *writer) writeWorksheet(sheet *Sheet) {
	if z.invalid() {
		return
	}
	name := z.createTarget("worksheets", fmt.Sprintf("%s.xml", sheet.Name))
	writer, err := z.writer.Create(name)
	if err != nil {
		z.err = err
		return
	}
	sw, err := writeSheet(writer)
	if err != nil {
		z.err = err
		return
	}
	if err := sw.WriteSheet(sheet); err != nil {
		z.err = err
	}
}

func (z *writer) writeWorkbook(f *File) {
	if z.invalid() {
		return
	}

	type xmlSheet struct {
		XMLName xml.Name   `xml:"sheet"`
		Id      string     `xml:"r:id,attr"`
		Name    string     `xml:"name,attr"`
		Index   int        `xml:"sheetId,attr"`
		State   SheetState `xml:"state,attr"`
	}

	root := struct {
		XMLName    xml.Name `xml:"workbook"`
		Xmlns      string   `xml:"xmlns,attr"`
		RelXmlns   string   `xml:"xmlns:r,attr"`
		Properties struct {
			Date int `xml:"date1904,attr"`
		} `xml:"workbookProperties"`
		Protection struct {
			Locked int `xml:"lockStructure,attr"`
		} `xml:"workbookProtection"`
		Views struct {
			View struct {
				activeTab int `xml:"activeTab,attr"`
			} `xml:"workbookView"`
		} `xml:"workbookViews"`
		Sheets []xmlSheet `xml:"sheets>sheet"`
	}{
		Xmlns:    typeMainUrl,
		RelXmlns: "http://schemas.openxmlformats.org/officeDocument/2006/relationships",
	}
	if f.locked {
		root.Protection.Locked++
	}
	if f.date1904 {
		root.Properties.Date++
	}
	for _, s := range f.sheets {
		s.Id = z.createFileID()
		s.Index = z.getFileIndex()
		xs := xmlSheet{
			Id:    s.Id,
			Index: s.Index,
			Name:  s.Name,
			State: s.State,
		}
		root.Sheets = append(root.Sheets, xs)
	}
	z.encodeXML(z.createTarget("workbook.xml"), root)
}

func (z *writer) encodeXML(name string, ptr any) {
	w, err := z.writer.Create(name)
	if err != nil {
		z.err = err
		return
	}
	if err := xml.NewEncoder(w).Encode(ptr); err != nil {
		z.err = fmt.Errorf("%w: fail to write data to %s", err, name)
	}
}

func (z *writer) createTarget(parts ...string) string {
	parts = append([]string{z.base}, parts...)
	return strings.Join(parts, "/")
}

func (z *writer) fromBase(target string) string {
	parts := strings.Split(target, "/")
	ix := slices.Index(parts, z.base)
	if ix < 0 {
		return target
	}
	return strings.Join(parts[ix+1:], "/")
}

func (z *writer) invalid() bool {
	return z.err != nil
}

func (z *writer) createFileID() string {
	z.lastUsedId++
	return fmt.Sprintf("rId%d", z.lastUsedId)
}

func (z *writer) getFileIndex() int {
	return z.lastUsedId - startIx
}

type sheetWriter struct {
	writer *sax.StreamWriter
}

func writeSheet(w io.Writer) (*sheetWriter, error) {
	sw, err := sax.Compact(w)
	if err != nil {
		return nil, err
	}
	sh := sheetWriter{
		writer: sw,
	}
	return &sh, nil
}

func (w *sheetWriter) WriteSheet(sheet *Sheet) error {
	wshName := sax.LocalName("worksheet")
	w.writer.Open(wshName, []sax.A{
		createNS("", typeMainUrl),
		createNS("r", "http://schemas.openxmlformats.org/officeDocument/2006/relationships"),
	})
	w.writer.Empty(sax.LocalName("dimension"), []sax.A{
		createAttr("ref", sheet.Bounding().String()),
	})
	if sheet.Protected != 0 {
		if err := w.writeProtection(sheet); err != nil {
			return err
		}
	}
	if err := w.writeRows(sheet); err != nil {
		return err
	}
	w.writer.Close(wshName)
	return w.writer.Flush()
}

func (w *sheetWriter) writeProtection(sheet *Sheet) error {
	var attrs []sax.A
	if sheet.Protected&ProtectedSheet != 0 {
		attrs = append(attrs, createAttr("sheet", "1"))
	}
	if sheet.Protected&ProtectedObjects != 0 {
		attrs = append(attrs, createAttr("objects", "1"))
	}
	if sheet.Protected&ProtectedScenarios != 0 {
		attrs = append(attrs, createAttr("scenarios", "1"))
	}
	if sheet.Protected&ProtectedFormatCells != 0 {
		attrs = append(attrs, createAttr("formatCells", "1"))
	}
	if sheet.Protected&ProtectedFormatColumns != 0 {
		attrs = append(attrs, createAttr("formatColumns", "1"))
	}
	if sheet.Protected&ProtectedDeleteColumns != 0 {
		attrs = append(attrs, createAttr("deleteColumns", "1"))
	}
	if sheet.Protected&ProtectedInsertColumns != 0 {
		attrs = append(attrs, createAttr("insertColumns", "1"))
	}
	if sheet.Protected&ProtectedDeleteRows != 0 {
		attrs = append(attrs, createAttr("deleteRows", "1"))
	}
	if sheet.Protected&ProtectedInsertRows != 0 {
		attrs = append(attrs, createAttr("insertRows", "1"))
	}
	if sheet.Protected&ProtectedSort != 0 {
		attrs = append(attrs, createAttr("sort", "1"))
	}
	return w.writer.Empty(sax.LocalName("sheetProtection"), attrs)
}

func (w *sheetWriter) writeRows(sheet *Sheet) error {
	var (
		dshName = sax.LocalName("sheetData")
		rowName = sax.LocalName("row")
	)
	w.writer.Open(dshName, nil)
	for _, r := range sheet.Rows {
		attrs := []sax.A{
			createAttr("r", strconv.FormatInt(r.Line, 10)),
		}
		if r.Hidden {
			attrs = append(attrs, createAttr("hidden", "1"))
		}
		w.writer.Open(rowName, attrs)
		for _, c := range r.Cells {
			if err := w.writeCell(c); err != nil {
				return err
			}
		}
		w.writer.Close(rowName)
	}
	w.writer.Close(dshName)
	return nil
}

func (w *sheetWriter) writeCell(cell *Cell) error {
	var (
		cellName = sax.LocalName("c")
		valName  = sax.LocalName("v")
		formName = sax.LocalName("f")
	)
	attrs := []sax.A{
		createAttr("r", cell.Position.Addr()),
	}
	if cell.Type != "" {
		attrs = append(attrs, createAttr("t", cell.Type))
	}
	w.writer.Open(cellName, attrs)
	if cell.Formula != nil {
		w.writer.Open(formName, nil)
		w.writer.Text(cell.Formula.String())
		w.writer.Close(formName)
	}
	w.writer.Open(valName, nil)
	w.writer.Text(cell.rawValue)
	w.writer.Close(valName)
	w.writer.Close(cellName)
	return nil
}

func createAttr(name, value string) sax.A {
	return sax.A{
		QName: sax.LocalName(name),
		Value: value,
	}
}

func createNS(name, value string) sax.A {
	qn := sax.LocalName("xmlns")
	if name != "" {
		qn = sax.QualifiedName(name, "xmlns")
	}
	return sax.A{
		QName: qn,
		Value: value,
	}
}
