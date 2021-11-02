package tests

import (
	"bytes"
	"strings"

	"github.com/olekukonko/tablewriter"
)

type TestCmdBuffers struct {
  InBuf *bytes.Buffer
  OutBuf *bytes.Buffer
  ErrBuf *bytes.Buffer
}

// TODO: Create a method to generate output using template
func GenerateTableOutputWithHeaders(headers []string, rows [][]string) string {
  tableBuf := &strings.Builder{}
  table := tablewriter.NewWriter(tableBuf)

  table.SetHeader(headers)
  table.SetColumnSeparator("")
  table.SetCenterSeparator("")
  table.SetRowSeparator("")
  table.SetAutoFormatHeaders(false)
  table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
  table.SetAutoWrapText(false)

  table.AppendBulk(rows)

  table.Render()
  return tableBuf.String()
}
