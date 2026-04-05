package eval

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"os"

	"github.com/midbel/dockit/formula/env"
	"github.com/midbel/dockit/formula/parse"
	"github.com/midbel/dockit/internal/slx"
	"github.com/midbel/dockit/value"
)

var (
	ErrEval     = errors.New("expression can not be evaluated")
	ErrCallable = errors.New("expression is not callable")
)

type Runnable interface {
	Run(parse.Expr) (value.Value, error)
}

type Engine struct {
	Stdout io.Writer
	Stderr io.Writer
	config *EngineConfig

	loaders map[string]Loader
}

func NewEngine() *Engine {
	e := Engine{
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		loaders: make(map[string]Loader),
		config:  NewConfig(),
	}
	e.RegisterLoader(".csv", CsvLoader())
	e.RegisterLoader(".xlsx", XlsxLoader())
	e.RegisterLoader(".ods", OdsLoader())
	e.RegisterLoader(".csv", CsvLoader())
	e.RegisterLoader(".log", LogLoader())
	return &e
}

func (e *Engine) SetContextDir(dir string) {
	if dir == "" {
		return
	}
	e.config.Set(slx.Make("context", "dir"), dir)
}

func (e *Engine) SetNumberFormat(format string) {
	if format == "" {
		return
	}
	e.config.Set(slx.Make("format", "number"), format)
}

func (e *Engine) SetDateFormat(format string) {
	if format == "" {
		return
	}
	e.config.Set(slx.Make("format", "date"), format)
}

func (e *Engine) SetPrintDebug(debug bool) {
	e.config.Set(slx.Make("print", "debug"), debug)
}

func (e *Engine) RegisterLoader(kind string, loader Loader) {
	e.loaders[kind] = loader
}

func (e *Engine) Exec(r io.Reader, environ *env.Environment) (value.Value, error) {
	ctx := NewEngineContext()
	ctx.loaders = maps.Clone(e.loaders)
	ctx.setEnv(environ)

	ps, err := e.bootstrap(r, ctx)
	if err != nil {
		return nil, err
	}
	switch ps.Mode() {
	case parse.ModeScript:
		expr, err := ps.Parse()
		if err != nil {
			return nil, err
		}
		eval := evalScript(ctx)
		return eval.Run(expr)
	default:
		return nil, fmt.Errorf("%s: unuspported mode", ps.Mode())
	}
}

func (e *Engine) bootstrap(r io.Reader, ctx *EngineContext) (*parse.Parser, error) {
	scan, err := parse.Scan(r, parse.ScanScript)
	if err != nil {
		return nil, err
	}
	ps, err := parse.NewParser(scan)
	if err != nil {
		return nil, err
	}
	entries, err := ps.ExtractConfigEntries()
	if err != nil {
		return nil, err
	}
	cfg := NewConfig()
	cfg.Merge(e.config)
	for _, e := range entries {
		cfg.Set(e.Path, e.Value)
	}
	if err := ctx.Configure(cfg); err != nil {
		return nil, err
	}
	return ps, nil
}
