package main

import (
	"fmt"
	"reflect"
	"test/test/package/packagedata"
)

func main() {
	obj := packagedata.Code1{
		Name:   "hello",
		Age:    100,
		Height: 1.2,
		Score:  []int{10, 30, 40},
		Child: []map[string]interface{}{
			map[string]interface{}{
				"Name": "hello",
			},
		},
	}
	fmt.Println(obj)

	conf := make(map[string]interface{})
	conf["1"] = obj
	// conf["-1"] = packagedata.Code2{}

	for code, instance := range conf {
		fmt.Println(instance)
		// fmt.Println("valueof", reflect.ValueOf(instance))
		st := reflect.TypeOf(instance)

		len := st.NumField()
		for i := 0; i < len; i++ {
			f := st.Field(i)
			switch f.Type.Kind() {
			case reflect.String:
				fmt.Println("is string")
			case reflect.Uint8:
				fmt.Println("is uint8")
			case reflect.Float32:
				fmt.Println("is float32")
			case reflect.Float64:
				fmt.Println("is float64")
			case reflect.Array:
				fmt.Println("is array")
			case reflect.Slice:
				fmt.Println("is slice")
				// vv := reflect.ValueOf(instance)
				// tt := reflect.TypeOf(vv)
				// switch reflect.MakeSlice(tt, 0, 0).Interface().(type) {
				// }
				// fmt.Println("===", vv.FieldByName(f.Name), vv.FieldByName(f.Name).Interface())
				// stvv := reflect.TypeOf(vv.FieldByName(f.Name))
				// lenvv := stvv.NumField()
				// fmt.Println(vv, lenvv)
				// for i := 0; i < lenvv; i++ {
				// 	ff := stvv.Field(i)
				// 	fmt.Println("name = ", ff.Name, ff.Type)
				// }
			case reflect.Struct:
				fmt.Println("is struct")
			}
			fmt.Println(code, f.Name, f.Type, f.Tag, f.Type.Name(), f.Type)
			// fmt.Printf("--%T\n", f.Type.Field(0))

			// fmt.Printf("%T %T %T\n", f.Name, f.Type, f.Type.Name())
			// switch f.Type.Name() {
			// case string:
			// 	fmt.Println("is string")
			// case int:
			// 	fmt.Println("is int")
			// }
		}
	}
}
