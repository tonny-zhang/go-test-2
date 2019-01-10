package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"path"

	"io/ioutil"

	"strings"

	"strconv"

	"github.com/tealeg/xlsx"
)

var dirConver, err = os.Getwd()

func errPrint(msg string) {
	fmt.Printf("!!\n\x1b[31;1m%s\x1b[0m\n!!\n\n", msg)
}
func getNumStr(numStr string) (string, bool) {
	i := strings.LastIndex(numStr, "%")
	isHavePercent := i > 0 && i == strings.Count(numStr, "")-1-1
	if isHavePercent {
		numStr = numStr[0:i]
	}
	return numStr, isHavePercent
}
func convert(excelFileName string) {
	excelFileName = strings.Replace(excelFileName, "\\", "/", -1)
	if strings.Index(path.Base(excelFileName), "~") == 0 {
		errPrint(excelFileName + " 文件名不合法，或不是一个完整的excel文件")
		return
	}
	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		fmt.Println(err)
		errPrint(excelFileName + " 文件内容错误，请检查最后几行空行")
		return
	}
	defer func() {
		if r := recover(); r != nil {
			errPrint(fmt.Sprintf("%s 解析错误 panic的内容%v\n", excelFileName, r))
		}
		// fmt.Printf("%s 解析错误，头列数不一致，请检查！\n", excelFileName)
	}()
	for _, sheet := range xlFile.Sheets {
		if strings.Index(sheet.Name, "_") != 0 || len(sheet.Rows) == 0 {
			continue
		}
		rows := sheet.Rows
		// fmt.Printf("rows = %d %s\n", len(rows), excelFileName)
		nameZH := rows[0].Cells
		types := rows[1].Cells
		nameEN := rows[2].Cells

		// lenZHCell := len(nameZH)
		lenEnCell := len(nameEN)
		var lenCellActual = 0
		var headerRow []map[string]string
		headerCache := make(map[string]bool)

		for i, cell := range nameZH {
			name := cell.String()
			if len(name) == 0 || i >= lenEnCell {
				break
			}
			enCell := nameEN[i]
			if nil != enCell {
				cellHeader := make(map[string]string)
				en := enCell.String()
				cellHeader["name"] = name
				cellHeader["en"] = en

				if headerCache[en] {
					errPrint(excelFileName + " 字段[" + en + "]重复")
					return
				}

				headerCache[en] = true
				typeCell := types[i]
				if nil != typeCell {
					t := types[i].String()
					cellHeader["type"] = t
				}
				headerRow = append(headerRow, cellHeader)
				lenCellActual++
			} else {
				break
			}

			// headerRow = append(headerRow, &headerCell{name, t, en})
		}
		// fmt.Println(headerRow)
		// b, err := json.Marshal(headerRow)
		// fmt.Println(string(b), err, len(b))
		// fmt.Printf("lenCellActual = %d\n", lenCellActual)

		// fmt.Printf("len_header = %d\n", lenEnCell)
		var data []map[string]interface{}
		lenReaded := 0
		for _, row := range rows[4:] {
			// if i != 32 {
			// 	// continue
			// }
			len := len(row.Cells)
			// fmt.Printf("len = %d\n", len)
			// if len > 0 && len <= lenZHCell {
			var dMap = make(map[string]interface{})
			var lenNull = 0
			var isEmpty = false
			// 强制读取，没有值时转换成默认值
			for index := 0; index < lenCellActual; index++ {
				var valStr = ""
				if index < len {
					valStr = row.Cells[index].String()
				}
				if index < lenCellActual {
					en := nameEN[index].String()
					t := types[index].String()
					t = strings.ToLower(t)
					if valStr == "" {
						if index == 0 {
							isEmpty = true
							break
						}
						lenNull++
					}

					// fmt.Printf("%d, %s, %s = %s\n", index, types[index].String(), en, valStr)
					numStr, isHavePercent := getNumStr(valStr)
					if t == "int" {
						valNumber, _ := strconv.Atoi(numStr)
						dMap[en] = valNumber
						if isHavePercent {
							dMap[en+"_isp"] = true
						}
					} else if t == "float" {
						valNumber, _ := strconv.ParseFloat(numStr, 64)
						dMap[en] = valNumber
						if isHavePercent {
							dMap[en+"_isp"] = true
						}
						// fmt.Printf("%s = %f\n", numStr, valNumber)
					} else if t == "bool" {
						dMap[en] = strings.ToUpper(valStr) == "T"
					} else {
						dMap[en] = valStr
					}

					// fmt.Printf("%s\t", d)
				}
			}
			// fmt.Printf("lenNull = %d, len = %d, lenZh = %d, lenCellActual = %d, %t, %v\n", lenNull, len, lenZHCell, lenCellActual, lenNull < len, dMap)
			// 过滤全空行
			if !isEmpty {
				lenReaded++
				data = append(data, dMap)
			}
		}

		// b1, err1 := json.Marshal(data)
		// fmt.Println(string(b1), err1, len(b1))
		// fmt.Printf("len = %d, datalen = %d, len_readed = %d, lenCellActual = %d\n", len(rows)-4, len(data), lenReaded, lenCellActual)
		result := make(map[string]interface{})
		result["header"] = headerRow
		result["root"] = data
		bResult, _ := json.Marshal(result)
		// fmt.Println(string(bResult), errResult, len(bResult))

		outputdir := path.Join(dirConver, "output")
		os.MkdirAll(outputdir, os.ModePerm)

		regPostfix := regexp.MustCompile("\\..+$")
		filenameNew := regPostfix.ReplaceAllString(path.Base(excelFileName), ".json")
		outfilename := path.Join(outputdir, filenameNew)

		f, err := os.Create(outfilename)
		defer f.Close()
		if err == nil {
			f.Write(bResult)
			fmt.Printf("%s save!\n", outfilename)
		} else {
			fmt.Println(err)
		}
	}
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
	// convert("E:\\source\\nodejs\\tool\\police\\data\\data\\skillEquip.xlsx")
	// walk("E:\\source\\nodejs\\tool\\police\\data\\data\\")
	// file := "E:\\source\\nodejs\\tool\\police\\data\\data\\activity.xlsx"
	// fmt.Println(file)
	// file = strings.Replace(file, "\\", "/", -1)
	// fmt.Println(file)
	// fmt.Println(path.Base(file))
	// reg := regexp.MustCompile("\\..+$")
	// fmt.Println(reg.ReplaceAllString(path.Base(file), ".json"))

	args := os.Args
	if len(args) > 1 {
		dirConver = args[1]
	}

	dirExcel := path.Join(dirConver, "data")
	if info, err := os.Stat(dirExcel); !os.IsNotExist(err) && info.IsDir() {
		walk(dirExcel)
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("\n\n回车退出...")
		reader.ReadByte()
		os.Exit(0)
	} else {
		errPrint("目录[" + dirConver + "]下没有用于存放excel文件的data目录")
	}
}
