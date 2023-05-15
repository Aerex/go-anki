package tests

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	prettytable "github.com/jedib0t/go-pretty/v6/table"
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
	table := prettytable.NewWriter()
	headerRow := prettytable.Row{}
	for _, header := range headers {
		headerRow = append(headerRow, header)
	}
	table.AppendHeader(headerRow)
	table.Style().Box.LeftSeparator = ""
	table.Style().Box.RightSeparator = ""
	table.Style().Box.MiddleSeparator = ""
	table.Style().Box.TopSeparator = ""
	table.Style().Box.BottomSeparator = ""

	//table.SetColumnSeparator("")
	//table.SetNoWhiteSpace(true)
	//table.SetCenterSeparator("")
	//table.SetHeader(headers)
	//table.SetRowSeparator("")
	//table.SetAutoFormatHeaders(false)
	//table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	//table.SetTablePadding("\t")
	//table.SetAutoWrapText(false)

	//table.AppendBulk(rows)

	return table.Render()
}
