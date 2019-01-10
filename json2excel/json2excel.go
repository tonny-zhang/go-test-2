package main
import (
	"strconv"
	"fmt"
	"os"
	"encoding/json"
	"io/ioutil"
	"path"
	"github.com/tealeg/xlsx"
	"regexp"
	"bufio"
)
var dirConver, _ = os.Getwd()
var outputdir = path.Join(dirConver, "output")
var exceldir = path.Join(dirConver, "excel")
var regPostfix = regexp.MustCompile("\\..+$")

func errPrint(msg string) {
	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", msg)
}
func convert(filepath string) {
	var xxx map[string][]map[string]interface{}

	// pathExcel := path.Join(outputdir, filename)
	
	filenameNew := regPostfix.ReplaceAllString(path.Base(filepath), ".xlsx")
	bytes, err := ioutil.ReadFile(filepath)
    if err != nil {
        fmt.Println("ReadFile: ", err.Error())
    }
	if err := json.Unmarshal(bytes, &xxx); err != nil {
        fmt.Println("Unmarshal: ", err.Error())
	}
	
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("_sheet")

	rowName := sheet.AddRow()
	rowType := sheet.AddRow();
	rowEn := sheet.AddRow();
	sheet.AddRow()

	for _, val := range xxx["header"] {
		cellName := rowName.AddCell()
		cellName.Value = val["name"].(string)

		cellEn := rowEn.AddCell()
		cellEn.Value = val["en"].(string)

		cellType := rowType.AddCell()
		cellType.Value = val["type"].(string)
	}

	for _, val := range xxx["root"] {
		row := sheet.AddRow();
		for _, rowHeader := range xxx["header"] {
			cell := row.AddCell()
			key := rowHeader["en"].(string)
			valType := rowHeader["type"].(string)

			if valType == "int" {
				cell.Value = strconv.Itoa(int(val[key].(float64)));
				if val[key+"_isp"] != nil && val[key+"_isp"].(bool) {
					cell.Value += "%"
				}
			} else if valType == "float" {
				cell.Value = strconv.FormatFloat(val[key].(float64), 'f', -1, 64);
				if val[key+"_isp"] != nil && val[key+"_isp"].(bool) {
					cell.Value += "%"
				}
			} else if valType == "bool" {
				if (val[key].(bool)) {
					cell.Value = "T"
				} else {
					cell.Value = "";
				}
			} else {
				cell.Value = val[key].(string)
			}
			
		}
	}
	pathExcelResult := path.Join(exceldir, filenameNew);
	err = file.Save(pathExcelResult)
	fmt.Printf("%s\n", pathExcelResult);
}

func walk(dir string) {
	files, err := ioutil.ReadDir(dir)
	if err == nil {
		for _, file := range files {
			filepath := path.Join(dir, file.Name())
			convert(filepath)
		}
	} else {
		fmt.Println(err)
	}
}

func main() {
	// fmt.Printf("hello %s\n", dirConver);
	os.MkdirAll(exceldir, os.ModePerm)

	args := os.Args
	if (len(args) > 1) {
		dirConver = args[1]
	}

	outputdir = path.Join(dirConver, "output")
	exceldir = path.Join(dirConver, "excel")

	if info, err := os.Stat(outputdir); !os.IsNotExist(err) && info.IsDir() {
		walk(outputdir)
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("\n\n回车退出...")
		reader.ReadByte()
		os.Exit(0)
	} else {
		errPrint("目录["+dirConver+"]下没有用于存放json文件的output目录")
	}
}