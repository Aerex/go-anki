package tests

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v2"
)

type TestCmdBuffers struct {
	InBuf  *bytes.Buffer
	OutBuf *bytes.Buffer
	ErrBuf *bytes.Buffer
}

// Return a yaml string from a json file
// TODO: Can't use this method because struct{} is not a given type
// Need a way to make it generic
func GenerateYamlOutputFromJSONFile(path string, model struct{}) (string, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(file, &model)
	if err != nil {
		return "", err
	}

	yamlData, err := yaml.Marshal(model)
	if err != nil {
		return "", err
	}

	return string(yamlData), nil
}

// TODO: Create a method to generate output using template
func GenerateTableOutputWithHeaders(headers []string, rows [][]string) string {
	tableBuf := &strings.Builder{}
	table := tablewriter.NewWriter(tableBuf)

	table.SetColumnSeparator("")
	table.SetNoWhiteSpace(true)
	table.SetCenterSeparator("")
	table.SetHeader(headers)
	table.SetRowSeparator("")
	table.SetAutoFormatHeaders(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetTablePadding("\t")
	table.SetAutoWrapText(false)

	table.AppendBulk(rows)

	table.Render()
	return tableBuf.String()
}
