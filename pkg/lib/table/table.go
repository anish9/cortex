/*
Copyright 2020 Cortex Labs, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package table

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cortexlabs/cortex/pkg/lib/console"
	"github.com/cortexlabs/cortex/pkg/lib/errors"
	"github.com/cortexlabs/cortex/pkg/lib/slices"
	s "github.com/cortexlabs/cortex/pkg/lib/strings"
)

type Table struct {
	Headers []Header
	Rows    [][]interface{}
	Spacing int // Spacing between rows. If 0 is provided, it defaults to 3.
}

type Header struct {
	Title    string
	MaxWidth int // Max width of the text (not including spacing). Items that are longer will be truncated to less than MaxWidth to fit the ellipses. If 0 is provided, it defaults to no max.
	MinWidth int // Min width of the text (not including spacing)
	Hidden   bool
}

func validate(t Table) error {
	numCols := len(t.Headers)

	if numCols < 1 {
		return errors.New("must have at least one column")
	}

	for _, header := range t.Headers {
		if header.MaxWidth != 0 && len(header.Title) > header.MaxWidth {
			return errors.New(fmt.Sprintf("header %s is wider than max width (%d)", header.Title, header.MaxWidth))
		}

		if header.MinWidth > header.MaxWidth {
			return errors.New(fmt.Sprintf("header %s has min width > max width (%d > %d)", header.Title, header.MinWidth, header.MaxWidth))
		}
	}

	for i, row := range t.Rows {
		if len(row) != numCols {
			return errors.New(fmt.Sprintf("row %d does not have the expected number of columns (%d)", i, numCols))
		}
	}

	return nil
}

// Prints the error message as a string (if there is an error)
func (t *Table) MustPrint() {
	fmt.Print(t.MustFormat())
}

// Return the error message as a string
func (t *Table) MustFormat() string {
	str, err := t.Format()
	if err != nil {
		return "error: " + errors.Message(err)
	}
	return str
}

func (t *Table) Format() (string, error) {
	if err := validate(*t); err != nil {
		return "", err
	}

	if t.Spacing <= 0 {
		t.Spacing = 3
	}

	colWidths := make([]int, len(t.Headers))
	for colNum, header := range t.Headers {
		colWidths[colNum] = len(header.Title)
	}

	rows := make([][]string, len(t.Rows))
	for rowNum, row := range t.Rows {
		rows[rowNum] = make([]string, len(row))
		for colNum, val := range row {
			strVal := s.ObjFlatNoQuotes(val)
			rows[rowNum][colNum] = strVal
			if len(strVal) > colWidths[colNum] {
				colWidths[colNum] = len(strVal)
			}
		}
	}

	maxColWidths := make([]int, len(t.Headers))
	for colNum, colWidth := range colWidths {
		if t.Headers[colNum].MaxWidth <= 0 {
			maxColWidths[colNum] = colWidth
		} else {
			maxColWidths[colNum] = slices.MinInt(colWidth, t.Headers[colNum].MaxWidth)
		}

		if maxColWidths[colNum] < t.Headers[colNum].MinWidth {
			maxColWidths[colNum] = t.Headers[colNum].MinWidth
		}
	}

	lastColIndex := len(t.Headers) - 1

	var headerStr string
	for colNum, header := range t.Headers {
		if header.Hidden {
			continue
		}
		headerStr += console.Bold(header.Title)
		if colNum != lastColIndex {
			headerStr += strings.Repeat(" ", maxColWidths[colNum]+t.Spacing-len(header.Title))
		}
	}

	ellipses := "..."
	rowStrs := make([]string, len(rows))
	for rowNum, row := range rows {
		var rowStr string
		for colNum, val := range row {
			if t.Headers[colNum].Hidden {
				continue
			}
			if len(val) > maxColWidths[colNum] {
				val = val[0:maxColWidths[colNum]]
				// Ensure at least one space after ellipses
				for len(val)+len(ellipses) > maxColWidths[colNum]+t.Spacing-1 {
					val = val[0 : len(val)-1]
				}
				val += ellipses
			}
			rowStr += val
			if colNum != lastColIndex {
				rowStr += strings.Repeat(" ", maxColWidths[colNum]+t.Spacing-len(val))
			}
		}
		rowStrs[rowNum] = rowStr
	}

	sort.Strings(rowStrs)

	return headerStr + "\n" + strings.Join(rowStrs, "\n") + "\n", nil
}
