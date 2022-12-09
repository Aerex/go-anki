package template

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aerex/go-anki/internal/utils"
	"github.com/aerex/go-anki/pkg/io"
	prettytable "github.com/jedib0t/go-pretty/v6/table"
	"github.com/olekukonko/tablewriter"
)

var alignmentMap = map[string]int{
	"left":   tablewriter.ALIGN_LEFT,
	"right":  tablewriter.ALIGN_RIGHT,
	"center": tablewriter.ALIGN_CENTER,
}

// TODO: Add the remaining styles
var tableStypeMap = map[string]prettytable.Style{
	"bold":             prettytable.StyleBold,
	"bright":           prettytable.StyleColoredBright,
	"dark":             prettytable.StyleColoredDark,
	"blackOnBlueWhite": prettytable.StyleColoredBlackOnBlueWhite,
	"rounded":          prettytable.StyleRounded,
}

func TableFuncMap(io *io.IO) map[string]interface{} {
	var table prettytable.Writer

	return map[string]interface{}{
		"row": func(entries ...interface{}) (string, error) {
			var row prettytable.Row
			for _, entry := range entries {
				switch entry := entry.(type) {
				case string:
					row = append(row, entry)
				case int:
					row = append(row, strconv.Itoa(entry))
				case bool:
					row = append(row, strconv.Itoa(utils.BoolToInt(entry)))
				}
			}
			if table != nil {
				table.AppendRow(row)
				table.AppendSeparator()
				return "", nil
			}
			return "", fmt.Errorf("missing table block")
		},
		"styles": func(styles ...string) (string, error) {
			if table == nil {
				return "", fmt.Errorf("missing {{table}} pipeline")
			}
			for _, style := range styles {
				styleEntries := strings.Split(style, "=")
				if len(styleEntries) == 1 {
					switch styleEntries[0] {
					case "border":
						table.Style().Options.DrawBorder = true
					case "headerLine":
					case "autoMergeCells":
					case "autoWrapText":
					case "autoFormatHeaders":
					case "noWhiteSpace":
					}
				} else if len(styleEntries) > 1 {
					name, value := styleEntries[0], styleEntries[1]
					switch name {
					case "theme":
						style, exists := tableStypeMap[value]
						if exists {
							table.SetStyle(style)
						}
					case "centerSep":
						table.Style().Box.MiddleSeparator = string(value)
					case "colSep":
						table.Style().Box.LeftSeparator = string(value)
						table.Style().Box.RightSeparator = string(value)
						//		//case "rowSep":
						//			//table.SetRowSeparator(string(value))
						//		//	if string(value) != "" {
						//		//		//table.SetRowLine(true)
						//		}
						//		case "headerAlignment":
						//			//align, found := alignmentMap[string(value)]
						//		//	if found {
						//		//		//table.SetHeaderAlignment(align)
						//		//	}
						//		case "colWidth":
						//			//intVal, err := strconv.Atoi(value)
						//			//if err != nil {
						//			//	return "", fmt.Errorf("expected a integer for colWidth but got %s", value)
						//			//}
						//			//table.SetColWidth(intVal)
					}
				} else {
					return "", fmt.Errorf("expected style prop in format `prop=value` or `prop` got %s instead", style)
				}
			}
			return "", nil
		},
		"headers": func(headerNames ...string) string {
			if table != nil {
				row := prettytable.Row{}
				for _, header := range headerNames {
					row = append(row, header)
				}
				table.AppendHeader(row)
			}
			return ""
		},
		"table": func() string {
			table = prettytable.NewWriter()
			table.SetOutputMirror(io.Output)
			table.SetStyle(prettytable.StyleDefault)
			return ""
		},
		"endtable": func() string {
			table.Render()
			return ""
		},
	}
}
