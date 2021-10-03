package template

import (
	"fmt"

	"github.com/aerex/anki-cli/pkg/io"
	"github.com/olekukonko/tablewriter"
)

func TableFuncMap(io *io.IO) map[string]interface{}{
  var table *tablewriter.Table

  return map[string]interface{} {
    "row": func(entries ...string) (string, error) {
      if table != nil {
        table.Append(entries)
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
