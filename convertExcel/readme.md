## 项目初始化步骤：
1. 把 `convert.go` 放在项目 src 目录下
1. 运行 `go get github.com/tealeg/xlsx` 安装解析 excel 的第三方依赖包
1. 运行 `go run convert.go` 可直接编译运行，也可以运行 `go build convert.go` 编译成二进制文件直接运行

## 使用方法：
自动行时会在当前可执行文件目录下寻找 `data` 目录，并遍历该目录下所有文件进行解析，并在当前目录下生成 `output` 文件夹，
并把生成的 json 文件同名放入 output 文件夹下。

## excel 格式说明
1. 暂时只支持导出 excel 中 sheet 名为 "_" 为 sheet
1. 第一行为字段中文描述
1. 第二行为字段数据类型
1. 第三行为字段英文名
1. 第四行为空行
1. 从第五行开始为相应数据

### excel 格式示例
ID | 名称  | 描述
----|------|----
int | string  | string
ID | name  | desc
 |   | 
1 | one  | 第一个
2 | two  | 第二个

### 转化后的 JSON 示例
```json
{
    "header": [{
        "name": "ID",
        "type": "int",
        "en": "ID"
    }, {
        "name": "名称",
        "type": "string",
        "en": "name"
    }, {
        "name": "描述",
        "type": "string",
        "en": "desc"
    }],
    "data": [{
        "ID": 1,
        "name": "one",
        "desc": "第一个"
    }, {
        "ID": 2,
        "name": "two",
        "desc": "第二个"
    }]
}
```