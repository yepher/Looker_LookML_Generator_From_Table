package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type InputFields struct {
	Column string
	Type   string
	Size   string
}

type TableInfo struct {
	Columns map[string]InputFields
}

func main() {
	tableInfo := flag.String("table", "", "Table info")
	lookerInput := flag.String("lookml", "", "Current LookML file")
	exludeSuffix := flag.String("suffix", "", "Column name suffix to exclude like '_c'")
	flag.Parse()

	if *tableInfo == "" {
		fmt.Printf("table is a required input file.")
		panic("table is a required input file.")
	}

	if *lookerInput == "" {
		fmt.Printf("LookML input file is required.")
		panic("LookML input file is required.")
	}

	///////
	// Read in Table CSV data
	///////
	filerc, err := os.Open(*tableInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer filerc.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(filerc)
	//contents := buf.String()

	r := csv.NewReader(strings.NewReader(buf.String()))
	r.Comma = ';'
	r.Comment = '#'

	columns := make(map[string]bool)
	table := make(map[string]InputFields)
	i := 0
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if i == 0 {
			// skip header row
			i++
			continue
		}
		//columns = append(columns, record[0])
		vals := strings.Split(record[0], ",")

		column := vals[0]
		columns[column] = true

		var columnDetail InputFields
		columnDetail.Column = column

		if len(vals) > 1 {
			columnDetail.Type = vals[1]
		}

		if len(vals) > 2 {
			columnDetail.Size = vals[2]
		}
		//fmt.Printf("Column Info: %v\n", columnDetail)
		table[column] = columnDetail
		i++
	}

	///////
	// Process LookML
	///////
	bytesRead, _ := ioutil.ReadFile(*lookerInput)
	fileContent := string(bytesRead)
	lines := strings.Split(fileContent, "\n")

	for _, line := range lines {
		if strings.Contains(line, "sql:") && strings.Contains(line, "${TABLE}.") {
			column := between(line, "${TABLE}.", ";")
			if column != "" {
				if strings.Contains(column, ",") {
					column = before(column, ",")
				}
				if strings.Contains(column, ":") {
					column = before(column, ":")
				}

				if strings.Contains(column, "=") {
					column = before(column, "=")
				}

				delete(columns, strings.TrimSpace(column))
				//fmt.Printf("%s\n", column)
			}
		}
	}

	//fmt.Printf("Remaining Columns: %v\n", columns)

	// Print remaining columns
	for k := range columns {
		colName, sqlType, _ := columnInfo(table, k)
		//fmt.Printf("Key Column  \"%s\", %s, %s\n", a, b, c)

		lookerType := lookmlType(sqlType)
		lookML := renderLookML(lookerType, colName, *exludeSuffix)

		fmt.Printf("%s\n", lookML)
	}

}

func between(value string, a string, b string) string {
	// Get substring between two strings.
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}
	posLast := strings.Index(value, b)
	if posLast == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(a)
	if posFirstAdjusted >= posLast {
		return ""
	}
	return value[posFirstAdjusted:posLast]
}

func before(value string, a string) string {
	pos := strings.Index(value, a)
	if pos == -1 {
		return ""
	}
	return value[0:pos]
}

func columnInfo(table map[string]InputFields, column string) (string, string, string) {
	columnInfo := table[column]
	//fmt.Printf("\tColumnInfo: %v\n", table)
	return columnInfo.Column, columnInfo.Type, columnInfo.Size
}

func lookmlType(sqlType string) string {
	//fmt.Printf("Looking up type for: %s\n", sqlType)
	if sqlType == "character varying" {
		return "string"
	} else if sqlType == "boolean" {
		return "yesno"
	} else if sqlType == "bigint" {
		return "number"
	} else if sqlType == "timestamp without time zone" {
		return "time"
	} else if sqlType == "double precision" {
		return "double"
	}

	return fmt.Sprintf("unknown(%s)", sqlType)
}

func renderLookML(colType string, colName string, suffix string) string {
	if colType == "time" {
		return renderTimeLookML(colType, colName, suffix)
	} else if colType == "time" {
		return renderTimeLookML(colType, colName, suffix)
	} else if colType == "double" {
		return renderDoubleLookML("number", colName, suffix)
	} else {
		return renderDefaultLookML(colType, colName, suffix)
	}
}

func renderDefaultLookML(colType string, colName string, suffix string) string {
	return "dimension: " + excludeSuffix(colName, suffix) + " {\n" +
		"    type: " + colType + "\n" +
		"    sql: ${TABLE}." + colName + " ;;\n" +
		"}\n"
}

func renderDoubleLookML(colType string, colName string, suffix string) string {
	return "dimension_group: " + excludeSuffix(colName, suffix) + " {\n" +
		"    type: " + colType + "\n" +
		"    timeframes: [raw, time, date, week, month, quarter, year]\n" +
		"    sql: ${TABLE}." + colName + "::decimal(20,7) ;;\n" +
		"}\n"
}

func renderTimeLookML(colType string, colName string, suffix string) string {
	return "dimension_group: " + excludeSuffix(colName, suffix) + " {\n" +
		"    type: " + colType + "\n" +
		"    timeframes: [raw, time, date, week, month, quarter, year]\n" +
		"    sql: ${TABLE}." + colName + " ;;\n" +
		"    convert_tz: no\n" +
		"}\n"
}

func excludeSuffix(name string, suffix string) string {
	if suffix == "" {
		return name
	}

	return before(name, suffix)
}
