package ods

import (
	"archive/zip"
	"compress/flate"
	"io"

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
	z.writeSettings()
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

func (z *writer) writeSettings() {
	if z.invalid() {
		return
	}
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
		tableName = sax.QualifiedName("table", "table")
		rowName   = sax.QualifiedName("table-row", "table")
		cellName  = sax.QualifiedName("table-cell", "table")
		textName  = sax.QualifiedName("p", "text")
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
		attrs := []sax.A{
			createAttr("name", "table", sh.Label),
		}
		if sh.Visible {
			attrs = append(attrs, createAttr("display", "table", "true"))
		}
		if sh.Locked {
			attrs = append(attrs, createAttr("protected", "table", "true"))
		}
		sx.Open(tableName, attrs)
		for _, row := range sh.rows {
			sx.Open(rowName, nil)
			for _, c := range row.Cells {
				var (
					typeName string
					val      = c.Value()
				)
				switch t := val.Type(); t {
				case value.TypeText:
					typeName = "string"
				case value.TypeNumber:
					typeName = "float"
				default:
					typeName = t
				}
				sx.Open(cellName, []sax.A{
					createAttr("value-type", "office", typeName),
					createAttr("value", "office", val.String()),
				})
				sx.Open(textName, nil)
				sx.Text(val.String())
				sx.Close(textName)
				sx.Close(cellName)
			}
			sx.Close(rowName)
		}
		sx.Close(tableName)
	}

	sx.Close(sheetName)
	sx.Close(bodyName)
	sx.Close(docName)
}

func (z *writer) invalid() bool {
	return z.err != nil
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
