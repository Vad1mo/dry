package appui

import (
	"fmt"
	"strconv"

	"github.com/docker/docker/api/types"
	units "github.com/docker/go-units"
	termui "github.com/gizak/termui"
	"github.com/moncho/dry/docker"
	"github.com/moncho/dry/ui"
	drytermui "github.com/moncho/dry/ui/termui"
)

//ContainerStatsRow is a Grid row showing runtime information about a container
type ContainerStatsRow struct {
	container *types.Container
	Name      *drytermui.ParColumn
	ID        *drytermui.ParColumn
	CPU       *drytermui.GaugeColumn
	Memory    *drytermui.GaugeColumn
	Net       *drytermui.ParColumn
	Block     *drytermui.ParColumn
	Pids      *drytermui.ParColumn
	X, Y      int
	Width     int
	Height    int
	columns   []termui.GridBufferer
}

//NewContainerStatsRow creates a ContainerStatsRow for the given container
func NewContainerStatsRow(s *docker.StatsChannel) *ContainerStatsRow {
	c := s.Container
	cf := docker.NewContainerFormatter(c, true)
	row := &ContainerStatsRow{
		container: c,
		Name:      drytermui.NewThemedParColumn(DryTheme, cf.Names()),
		ID:        drytermui.NewThemedParColumn(DryTheme, cf.ID()),
		CPU:       drytermui.NewThemedGaugeColumn(DryTheme),
		Memory:    drytermui.NewThemedGaugeColumn(DryTheme),
		Net:       drytermui.NewThemedParColumn(DryTheme, "-"),
		Block:     drytermui.NewThemedParColumn(DryTheme, "-"),
		Pids:      drytermui.NewThemedParColumn(DryTheme, "-"),

		Height: 1,
	}
	//Columns are rendered following the slice order
	row.columns = []termui.GridBufferer{
		row.ID,
		row.Name,
		row.CPU,
		row.Memory,
		row.Net,
		row.Block,
		row.Pids,
	}
	if docker.IsContainerRunning(c) {
		go func() {
			for stat := range s.Stats {
				row.setNet(stat.NetworkRx, stat.NetworkTx)
				row.setCPU(stat.CPUPercentage)
				row.setMem(stat.Memory, stat.MemoryLimit, stat.MemoryPercentage)
				row.setBlockIO(stat.BlockRead, stat.BlockWrite)
				row.setPids(stat.PidsCurrent)
			}
		}()
	} else {
		row.markAsNotRunning()
	}
	return row
}

//Reset resets row content
func (row *ContainerStatsRow) Reset() {
	row.CPU.Reset()
	row.Memory.Reset()
	row.Net.Reset()
	row.Pids.Reset()
	row.Block.Reset()
}

//GetHeight returns this ContainerStatsRow heigth
func (row *ContainerStatsRow) GetHeight() int {
	return row.Height
}

//SetX sets the x position of this ContainerStatsRow
func (row *ContainerStatsRow) SetX(x int) {
	row.X = x
}

//SetY sets the y position of this ContainerStatsRow
func (row *ContainerStatsRow) SetY(y int) {
	if y == row.Y {
		return
	}
	for _, col := range row.columns {
		col.SetY(y)
	}
	row.Y = y
}

//SetWidth sets the width of this ContainerStatsRow
func (row *ContainerStatsRow) SetWidth(width int) {
	if width == row.Width {
		return
	}
	row.Width = width
	x := row.X
	rw := calcItemWidth(width, len(row.columns))
	for _, col := range row.columns {
		col.SetX(x)
		col.SetWidth(rw)
		x += rw + columnSpacing
	}
}

//Buffer returns this ContainerStatsRow data as a termui.Buffer
func (row *ContainerStatsRow) Buffer() termui.Buffer {
	buf := termui.NewBuffer()

	buf.Merge(row.ID.Buffer())
	buf.Merge(row.Name.Buffer())
	buf.Merge(row.CPU.Buffer())
	buf.Merge(row.Memory.Buffer())
	buf.Merge(row.Net.Buffer())
	buf.Merge(row.Block.Buffer())
	buf.Merge(row.Pids.Buffer())

	return buf
}

func (row *ContainerStatsRow) setNet(rx float64, tx float64) {
	row.Net.Text = fmt.Sprintf("%s / %s", units.BytesSize(rx), units.BytesSize(tx))
}

func (row *ContainerStatsRow) setBlockIO(read float64, write float64) {
	row.Block.Text = fmt.Sprintf("%s / %s", units.BytesSize(read), units.BytesSize(write))
}
func (row *ContainerStatsRow) setPids(pids uint64) {
	row.Pids.Text = strconv.Itoa(int(pids))
}

func (row *ContainerStatsRow) setCPU(val float64) {
	row.CPU.Label = fmt.Sprintf("%.2f%%", val)
	cpu := int(val)
	if cpu < 5 {
		cpu = 5
	} else if cpu > 100 {
		cpu = 100
	}
	row.CPU.Percent = cpu
	row.CPU.BarColor = percentileToColor(cpu)
}

func (row *ContainerStatsRow) setMem(val float64, limit float64, percent float64) {
	row.Memory.Label = fmt.Sprintf("%s / %s", units.BytesSize(val), units.BytesSize(limit))
	mem := int(percent)
	if mem < 5 {
		mem = 5
	} else if mem > 100 {
		mem = 100
	}
	row.Memory.Percent = mem
	row.Memory.BarColor = percentileToColor(mem)
}

//markAsNotRunning
func (row *ContainerStatsRow) markAsNotRunning() {
	c := termui.Attribute(ui.Color244)
	row.Name.TextFgColor = c
	row.ID.TextFgColor = c
	row.CPU.PercentColor = c
	row.CPU.Label = "-"
	row.Memory.PercentColor = c
	row.Memory.Label = "-"
	row.Net.TextFgColor = c
}

func percentileToColor(n int) termui.Attribute {
	c := ui.Color23
	if n > 90 {
		c = ui.Color161
	} else if n > 70 {
		c = ui.Color131
	}
	return termui.Attribute(c)
}
