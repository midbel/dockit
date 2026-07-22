package grid

import (
	"fmt"
	"iter"
	"slices"
)

type Link struct {
	UsedBy    []Cell
	DependsOn []Cell
}

func (k *Link) AddUsedBy(cell Cell) {
	ok := slices.ContainsFunc(k.UsedBy, func(c Cell) bool {
		return c.Id() == cell.Id()
	})
	if !ok {
		k.UsedBy = append(k.UsedBy, cell)
	}
}

func (k *Link) AddDependsOn(cell Cell) {
	ok := slices.ContainsFunc(k.DependsOn, func(c Cell) bool {
		return c.Id() == cell.Id()
	})
	if !ok {
		k.DependsOn = append(k.DependsOn, cell)
	}
}

func BuildGraph(file File) error {
	if err := ClearGraph(file); err != nil {
		return err
	}
	for _, sh := range file.Sheets() {
		if err := buildSheetGraph(file, sh); err != nil {
			return err
		}
	}
	return nil
}

type dependsOnCell interface {
	AddDependency(Cell)
}

type usedByCell interface {
	AddDependent(Cell)
}

type mutableCell interface {
	UsedBy() []Cell
	Dirty() bool
	MarkDirty()
}

func MarkDirty(cell Cell) {
	mc, ok := cell.(mutableCell)
	if !ok {
		return
	}
	if mc.Dirty() {
		return
	}
	mc.MarkDirty()
	for _, c := range mc.UsedBy() {
		MarkDirty(c)
	}
}

func LinkCells(source, target Cell) error {
	var (
		d, ok1 = source.(dependsOnCell)
		u, ok2 = target.(usedByCell)
	)
	if ok1 && ok2 {
		d.AddDependency(target)
		u.AddDependent(source)
	} else {
		return fmt.Errorf("dependency can not be created")
	}
	return nil
}

func ClearGraph(file File) error {
	return nil
}

func buildSheetGraph(file File, view View) error {
	for c := range iterCellsFromView(view) {
		f := c.Formula()
		if f == nil {
			continue
		}
		for _, pos := range Dependencies(f) {
			var (
				sh  View
				err error
			)
			if pos.Sheet == "" {
				sh = view
			} else {
				sh, err = file.Sheet(pos.Sheet)
			}
			if err != nil {
				return err
			}
			cell, err := sh.Cell(pos)
			if err != nil {
				return err
			}
			LinkCells(c, cell)
		}
	}
	return nil
}

func iterCellsFromView(view View) iter.Seq[Cell] {
	it := func(yield func(Cell) bool) {
		bd := view.Bounds()
		for pos := range bd.Positions() {
			cell, err := view.Cell(pos)
			if err != nil {
				continue
			}
			ok := yield(cell)
			if !ok {
				return
			}
		}
	}
	return it
}
