package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"path"
	"regexp"
	"strconv"
	"errors"

	"github.com/nfnt/resize"
)

type root struct {
	Width  string
	Height string
	Res    string
}
type scene struct {
	Root []root
}

var regData, _ = regexp.Compile(`m_IndexBuffer:\s+(\w+?)\s[\s\S]*_typelessdata:\s+(\w+?)\s`)
var dirCurrent, err = os.Getwd()
const scale = 1;

func errPrint(msg string) {
	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", msg)
}
func parseMesh(meshPath string) ([]uint16, [][]float32, error) {
	contents, err := ioutil.ReadFile(meshPath)

	if nil == err {
		m := regData.FindStringSubmatch(string(contents))
		if len(m) > 0 {
			var hexIndexData = m[1]
			var hexVerticesData = m[2]

			indexData, _ := hex.DecodeString(hexIndexData)
			verticesData, _ := hex.DecodeString(hexVerticesData)

			lenIndexData := len(indexData)
			lenVerticesData := len(verticesData) / 32

			indexArr := make([]uint16, lenIndexData/2)
			for i := 0; i < lenIndexData; i += 2 {
				val := binary.LittleEndian.Uint16(indexData[i : i+2])
				// fmt.Println(i/2, val)
				indexArr[i/2] = val
			}

			verticesArr := make([][]float32, lenVerticesData)
			for i := 0; i < lenVerticesData; i++ {
				_indexRead := i * 32
				xBytes := binary.LittleEndian.Uint32(verticesData[_indexRead : _indexRead+4])
				yBytes := binary.LittleEndian.Uint32(verticesData[_indexRead+4 : _indexRead+8])
				xFloat := math.Float32frombits(xBytes)
				yFloat := math.Float32frombits(yBytes)

				// fmt.Println(xFloat, yFloat)
				verticesArr[i] = []float32{xFloat, yFloat}
			}

			return indexArr, verticesArr, nil
		}
	} else {
		fmt.Println("read data error")
	}
	return nil, nil, nil
}
func getImgSize(imgPath string) (int, int, error) {
	file, err := os.Open(imgPath)
	if err != nil {
		return 0, 0, err
	}
	img, _, errImg := image.DecodeConfig(file)
	if errImg != nil {
		return 0, 0, errImg
	}
	return img.Width, img.Height, nil
}
func getPoint(indexArr []uint16, verticesArr [][]float32, width, height int) func(index uint16) (x, y float64) {
	widthHalfFloat := float32(width / 2)
	heightHalfFloat := float32(height / 2)
	return func(index uint16) (x, y float64) {
		v := verticesArr[index]
		return float64(int(v[0]*100 + widthHalfFloat)), float64(int(heightHalfFloat - v[1]*100))
	}
}
func parseWithSize(meshPath string, width, height int) ([]byte, image.Image, error) {
	indexArr, verticesArr, err := parseMesh(meshPath)

	if err != nil {
		return nil, nil, err
	}
	fnGetPoint := getPoint(indexArr, verticesArr, width, height)
	lenIndex := len(indexArr)
	fmt.Printf("width = %d, height = %d\n", width, height)

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	c := color.RGBA{255, 0, 0, 255}
	for i := 0; i < lenIndex; i += 3 {
		points := make([][]float64, 3)
		x1, y1 := fnGetPoint(indexArr[i])
		points[0] = []float64{x1, y1}
		xMin := x1
		xMax := x1
		yMin := y1
		yMax := y1
		x2, y2 := fnGetPoint(indexArr[i+1])
		points[1] = []float64{x2, y2}
		xMin = math.Min(xMin, x2)
		xMax = math.Max(xMax, x2)
		yMin = math.Min(yMin, y2)
		yMax = math.Max(yMax, y2)
		x3, y3 := fnGetPoint(indexArr[i+2])
		points[2] = []float64{x3, y3}
		xMin = math.Min(xMin, x3)
		xMax = math.Max(xMax, x3)
		yMin = math.Min(yMin, y3)
		yMax = math.Max(yMax, y3)

		for x := xMin; x < xMax; x++ {
			for y := yMin; y < yMax; y++ {
				if isPointInTriangle(points, x, y) {
					img.SetRGBA(int(x), int(y), c)
				}
			}
		}
	}

	widthMini := uint(math.Ceil(float64(width/scale)))
	imgMini := resize.Resize(widthMini, 0, img, resize.NearestNeighbor)
	widthNew := imgMini.Bounds().Dx()
	heightNew := imgMini.Bounds().Dy()
	fmt.Printf("widthMin = %d, heightMin = %d, scale = %d\n", widthNew, heightNew, scale)

	imgMiniRGBA, ok := imgMini.(*image.RGBA)
	if (!ok) {
		fmt.Printf("%s\n", "convert to rgba error");
		return nil, nil, errors.New("convert to rgba error")
	}
	lenResult := int(math.Ceil(float64(widthNew * heightNew / 8)))
	resultBytes := make([]byte, lenResult + 4 + 4 + 4)
	binary.LittleEndian.PutUint32(resultBytes, uint32(widthNew))
	binary.LittleEndian.PutUint32(resultBytes[4:], uint32(heightNew))
	binary.LittleEndian.PutUint32(resultBytes[8:], uint32(scale))
	dataWritedIndex := 12
	lenWrited := 0
	var dataWrited byte
	resultArr := make([][]int, widthNew)
	for i := 0; i < widthNew; i++ {
		rowArr := make([]int, heightNew)
		for j := 0; j < heightNew; j++ {
			var v int
			if imgMiniRGBA.RGBAAt(i, j).A > 0 {
				v = 1
			}
			rowArr[j] = v
			if lenWrited == 8 {
				resultBytes[dataWritedIndex] = dataWrited
				dataWritedIndex++
				dataWrited = 0
				lenWrited = 0
			}
			if v == 1 {
				dataWrited += byte(v << uint(7-lenWrited))
			}
			lenWrited++
		}
		resultArr[i] = rowArr
	}
	// fmt.Println(lenWrited, dataWrited, dataWritedIndex, lenResult)
	if lenWrited > 0 && dataWritedIndex < lenResult {
		resultBytes[dataWritedIndex] = dataWrited
	}

	
	// dirMeshImg := path.Join(dirCurrent, "meshImg")
	// imgfile, _ := os.Create(path.Join(dirMeshImg, "1.png"))
	// defer imgfile.Close()
	// png.Encode(imgfile, imgMini)

	fmt.Println("");
	return resultBytes, imgMini, nil
}
func parse(meshPath, imgPath string) ([]byte, image.Image, error) {
	width, height, errImg := getImgSize(imgPath)

	if errImg != nil {
		return nil, nil, errImg
	}
	return parseWithSize(meshPath, width, height)
}
func isPointInTriangle(points [][]float64, x, y float64) bool {
	lenPoints := len(points)
	j := lenPoints - 1
	inside := false
	for i := 0; i < lenPoints; i++ {
		xi := points[i][0]
		yi := points[i][1]
		xj := points[j][0]
		yj := points[j][1]

		if x == xi && y == yi || (x == xj && y == yj) {
			return true
		}
		if x == xi && x == xj && (y-yi)*(y-yj) <= 0 || (y == yj && y == yi && (x-xi)*(x-xj) <= 0) {
			return true
		}
		intersect := ((yi > y) != (yj > y)) && (x <= (xj-xi)*(y-yi)/(yj-yi)+xi)
		if intersect {
			inside = !inside
		}
		j = i
	}
	return inside
}
func main() {
	// parse("D:/source/unity3d/parseMesh/Assets/Resources/NoviceTrain.asset", "E:/source/nodejs/gameMesh/data/NoviceTrain.jpg")

	dirData := path.Join(dirCurrent, "data")
	if info, err := os.Stat(dirData); !os.IsNotExist(err) && info.IsDir() {
		jsonPath := path.Join(dirData, "scene.json")
		if info, err := os.Stat(jsonPath); !os.IsNotExist(err) && !info.IsDir() {
			dirMeshImg := path.Join(dirCurrent, "meshImg")
			dirMeshData := path.Join(dirCurrent, "meshData")
			os.MkdirAll(dirMeshImg, 0777)
			os.MkdirAll(dirMeshData, 0777)
			contents, err := ioutil.ReadFile(jsonPath)
			if err == nil {
				sceneData := &scene{}
				if err := json.Unmarshal(contents, sceneData); nil == err {

					for index, v := range sceneData.Root {
						// if index < 51 {
						// 	continue
						// }
						res := v.Res
						// if res != "NoviceTrain" {
						// 	continue
						// }
						fmt.Println(index, res)
						width, _ := strconv.ParseInt(v.Width, 10, 32)
						height, _ := strconv.ParseInt(v.Height, 10, 32)
						pathAsset := path.Join(dirData, res+".asset")
						pathDat := path.Join(dirMeshData, res+".dat")
						pathImg := path.Join(dirMeshImg, res+".png")
						if infoAsset, err := os.Stat(pathAsset); !os.IsNotExist(err) {
							infoDat, errDat := os.Stat(pathDat); 
							infoImg, errImg := os.Stat(pathImg);
							// 已经生成的数据文件或图片在源文件修改时间后可以说明源文件没有修改，没有修改时不用重复生成，减小生成的总时间
							if (os.IsNotExist(errDat) || infoAsset.ModTime().Unix() > infoDat.ModTime().Unix()) ||
								(os.IsNotExist(errImg) || infoAsset.ModTime().Unix() > infoImg.ModTime().Unix()) {
									resultByte, img, err := parseWithSize(pathAsset, int(width), int(height))
									if nil == err {
										fResult, _ := os.Create(pathDat)
										defer fResult.Close()
										fResult.Write(resultByte)
			
										imgfile, _ := os.Create(pathImg)
										defer imgfile.Close()
										png.Encode(imgfile, img)
										fmt.Println(res + "------")
									} else {
										fmt.Println(err)
										errPrint(res + "处理出现错误")
									}
							} else {
								errPrint(res + "文件没有更改");
							}
						} else {
							errPrint(res + "不存在")
						}						
					}
				} else {
					fmt.Println(err)
					errPrint("解析scene.json出现错误")
				}
			} else {
				errPrint("读取scene.json时出现错误")
			}
		} else {
			errPrint("data目录下没有scene.json")
		}
	} else {
		errPrint("当前目录下没有用于存放scene.json及mesh文件的data目录")
	}
}
