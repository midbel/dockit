package ods

import (
	"archive/zip"
	"compress/flate"
	"io"
	"strconv"

	sax "github.com/midbel/codecs/xml"
	"github.com/midbel/dockit/value"
)

type writer struct {
	writer *zip.Writer
	err    error
}

func writeFile(w io.Writer) (*writer, error) {
	z := writer{
		writer: zip.NewWriter(w),
	}
	z.writer.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestCompression)
	})
	return &z, nil
}

func (z *writer) WriteFile(file *File) error {
	z.writeMime()
	z.writeStyle()
	z.writeContent(file)
	z.writeSettings(file)
	z.writeMeta()
	z.writeManifest()
	return z.err
}

func (w *writer) Close() error {
	return w.writer.Close()
}

func (z *writer) writeMime() {
	if z.invalid() {
		return
	}
	w, err := z.writer.Create("mimetype")
	if err != nil {
		z.err = err
		return
	}
	_, z.err = io.WriteString(w, mimeODS)
}

func (z *writer) writeStyle() {
	if z.invalid() {
		return
	}
}

func (z *writer) writeMeta() {
	if z.invalid() {
		return
	}
}
func (z *writer) writeManifest() {
	if z.invalid() {
		return
	}
	w, err := z.writer.Create("META-INF/manifest.xml")
	if err != nil {
		z.err = err
		return
	}

	sx, err := sax.Compact(w)
	if err != nil {
		z.err = err
		return
	}
	defer sx.Flush()

	var (
		manifestName = sax.QualifiedName("manifest", "manifest")
		entryName    = sax.QualifiedName("file-entry", "manifest")
	)
	sx.Open(manifestName, []sax.A{
		createNS("manifest", manifestNS),
		createAttr("version", "manifest", officeVersion),
	})
	sx.Empty(entryName, []sax.A{
		createAttr("full-path", "manifest", "/"),
		createAttr("media-type", "manifest", mimeODS),
	})
	sx.Empty(entryName, []sax.A{
		createAttr("full-path", "manifest", "content.xml"),
		createAttr("media-type", "manifest", mimeXML),
	})
	sx.Close(manifestName)
}

func (z *writer) writeSettings(file *File) {
	if z.invalid() {
		return
	}
	w, err := z.writer.Create("settings.xml")
	if err != nil {
		z.err = err
		return
	}
	sx, err := sax.Compact(w)
	if err != nil {
		z.err = err
		return
	}
	defer sx.Flush()

	activeSheet, _ := file.activeSheet()

	var (
		rootName    = sax.QualifiedName("document-settings", "office")
		settingName = sax.QualifiedName("settings", "office")
		setName     = sax.QualifiedName("config-item-set", "config")
		itemName    = sax.QualifiedName("config-item", "config")
	)

	sx.Open(rootName, []sax.A{
		createNS("office", officeNS),
		createNS("config", configNS),
		createAttr("version", "office", officeVersion),
	})
	sx.Open(settingName, nil)
	sx.Open(setName, []sax.A{
		createAttr("name", "config", "ooo:view-settings"),
	})
	if activeSheet != nil {
		sx.Open(itemName, []sax.A{
			createAttr("name", "config", "ActiveTable"),
			createAttr("type", "config", "string"),
		})
		sx.Text(activeSheet.Label)
		sx.Close(itemName)
	}
	sx.Close(setName)
	sx.Close(settingName)
	sx.Close(rootName)
}

func (z *writer) writeContent(file *File) {
	if z.invalid() {
		return
	}
	w, err := z.writer.Create("content.xml")
	if err != nil {
		z.err = err
		return
	}
	sx, err := sax.Compact(w)
	if err != nil {
		z.err = err
		return
	}
	defer sx.Flush()

	var (
		docName   = sax.QualifiedName("document-content", "office")
		bodyName  = sax.QualifiedName("body", "office")
		sheetName = sax.QualifiedName("spreadsheet", "office")
	)

	sx.Open(docName, []sax.A{
		createNS("office", officeNS),
		createNS("table", tableNS),
		createNS("text", textNS),
		createAttr("version", "office", officeVersion),
	})
	sx.Open(bodyName, nil)
	sx.Open(sheetName, nil)

	for _, sh := range file.sheets {
		w := writeSheet(sx)
		if err := w.WriteSheet(sh); err != nil {
			z.err = err
			break
		}
	}
	sx.Close(sheetName)
	sx.Close(bodyName)
	sx.Close(docName)
}

func (z *writer) invalid() bool {
	return z.err != nil
}

type sheetWriter struct {
	writer *sax.StreamWriter
}

func writeSheet(writer *sax.StreamWriter) *sheetWriter {
	sh := sheetWriter{
		writer: writer,
	}
	return &sh
}

func (w *sheetWriter) WriteSheet(sheet *Sheet) error {
	tableName := sax.QualifiedName("table", "table")
	attrs := []sax.A{
		createAttr("name", "table", sheet.Label),
	}
	if sheet.Visible {
		attrs = append(attrs, createAttr("display", "table", "true"))
	}
	if sheet.Locked {
		attrs = append(attrs, createAttr("protected", "table", "true"))
	}
	w.writer.Open(tableName, attrs)
	for i, row := range sheet.rows {
		var diff int64
		if i > 0 {
			diff = row.Line - sheet.rows[i-1].Line
		}
		w.writeRow(row, diff)
	}
	w.writer.Close(tableName)
	return nil
}

func (w *sheetWriter) writeRow(row *row, delta int64) {
	var (
		rowName  = sax.QualifiedName("table-row", "table")
		cellName = sax.QualifiedName("table-cell", "table")
		textName = sax.QualifiedName("p", "text")
		attrs    []sax.A
	)
	if delta > 1 {
		attrs = append(attrs, createAttr("number-rows-repeated", "table", strconv.FormatInt(delta, 10)))
	}
	if row.Len() == 0 {
		w.writer.Empty(rowName, attrs)
		return
	}

	w.writer.Open(rowName, attrs)

	var lastCol int64
	for i := 0; i < len(row.Cells); {
		var (
			cell = row.Cells[i]
			val  = cell.Value()
		)
		if diff := cell.Column - lastCol; diff > 1 {
			w.writer.Empty(cellName, []sax.A{
				createAttr("number-columns-repeated", "table", strconv.FormatInt(diff-1, 10)),
			})
		}
		repeat := 1
		for j := i + 1; j < len(row.Cells); j++ {
			next := row.Cells[j]
			if next.Column-row.Cells[j-1].Column > 1 {
				break
			}

			if !cell.Equal(next) {
				break
			}
			repeat++
		}
		attrs := []sax.A{
			createAttr("value-type", "office", getTypeFromValue(val)),
			createAttr("value", "office", val.String()),
		}
		if repeat > 1 {
			attrs = append(attrs, createAttr("number-columns-repeated", "table", strconv.Itoa(repeat)))
		}
		w.writer.Open(cellName, attrs)
		w.writer.Open(textName, nil)
		w.writer.Text(val.String())
		w.writer.Close(textName)
		w.writer.Close(cellName)

		i += repeat
		lastCol = cell.Column + int64(repeat) - 1
	}
	w.writer.Close(rowName)
}

func getTypeFromValue(val value.Value) string {
	var ret string
	switch t := val.Type(); t {
	case value.TypeText:
		ret = "string"
	case value.TypeNumber:
		ret = "float"
	default:
		ret = t
	}
	return ret
}

func createAttr(name, space, value string) sax.A {
	return sax.A{
		QName: sax.QualifiedName(name, space),
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
