package template

import (
	"fmt"
	"strconv"

	"github.com/aerex/anki-cli/internal/utils"
	"github.com/aerex/anki-cli/pkg/io"
	"github.com/olekukonko/tablewriter"
)

func TableFuncMap(io *io.IO) map[string]interface{}{
  var table *tablewriter.Table

  return map[string]interface{} {
    "row": func(entries ...interface{}) (string, error) {
      var row []string
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
        table.Append(row)
        return "", nil
      }
      return "", fmt.Errorf("missing table block")
    },
    "headers": func(headerNames ...string) string {
      if table != nil {
        table.SetHeader(headerNames)
      }
      return ""
    },
    "table": func() string {
      table = tablewriter.NewWriter(io.Output)
      // TODO: Move settings into seperate block to customize
      table.SetColumnSeparator("")
      table.SetCenterSeparator("")
      table.SetRowSeparator("")
      table.SetAutoFormatHeaders(false)
      table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
      table.SetAutoWrapText(false)
      return ""
    },
    "endtable": func() string {
      table.Render()
      return ""
    },
  }
}
