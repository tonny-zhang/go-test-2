package main

import (
	"database/sql"
	"fmt"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
)

func check(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
func query(sqlStr string) ([]interface{}, error) {
	db, err := sql.Open("mysql", "root:12345678@tcp(localhost:3306)/xy3")

	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(sqlStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make([]interface{}, 0)
	for rows.Next() {
		columns, _ := rows.Columns()
		record := make(map[string]interface{})
		scanArgs := make([]interface{}, len(columns))
		values := make([]interface{}, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		rows.Scan(scanArgs...)
		fmt.Println(len(columns), values)
		for i, col := range values {
			if col != nil {
				record[columns[i]] = string(col.([]byte))
			}
		}
		data = append(data, record)
	}
	return data, nil
}
func query1(sqlStr string) ([]rowID, error) {
	db, err := sql.Open("mysql", "root:12345678@tcp(localhost:3306)/xy3")

	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(sqlStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := make([]rowID, 0)
	for rows.Next() {
		var rowData rowID
		rows.Scan(&rowData.ID)
		values = append(values, rowData)
	}
	return values, nil
}
func query2(sqlStr string, fnScan func(*sql.Rows)) error {
	db, err := sql.Open("mysql", "root:12345678@tcp(localhost:3306)/xy3")

	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Query(sqlStr)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		fnScan(rows)
	}
	return nil
}
func mysqlTest() {
	db, err := sql.Open("mysql", "root:12345678@tcp(localhost:3306)/xy3")

	check(err)
	defer db.Close()

	rows, err := db.Query("SELECT * from role where time_login < '2019-03-25 00:00:00'")
	check(err)
	defer rows.Close()
	for rows.Next() {
		cTypes, _ := rows.ColumnTypes()
		columns, _ := rows.Columns()

		for i, v := range cTypes {
			fmt.Println(i, v.Name(), v.ScanType())
		}
		fmt.Println("++++++")
		scanArgs := make([]interface{}, len(columns))
		values := make([]interface{}, len(columns))

		for i := range values {
			scanArgs[i] = &values[i]
		}

		fmt.Println("scanArgs", scanArgs)
		//将数据保存到 record 字典
		err = rows.Scan(scanArgs...)
		check(err)
		record := make(map[string]string)
		for i, col := range values {
			if col != nil {
				record[columns[i]] = string(col.([]byte))
			}
		}

		fmt.Println(record)
	}
}
func redisTest() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	dataText := make(map[string]interface{})
	dataText["id"] = 1
	dataText["name"] = "hello"
	client.HMSet("test", dataText)
	val, err := client.HGetAll("test").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("key", val)

	valDele, err := client.Del("test").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("val_dele", valDele)
}

type rowID struct {
	ID int
}

func main() {
	// mysqlTest()

	// result, _ := query("SELECT id from role where time_login < '2019-03-25 00:00:00'")
	// fmt.Println(result)

	// result1, _ := query1("SELECT id from role where time_login < '2019-03-25 00:00:00'")
	// fmt.Println(result1)

	// result2 := make([]rowID, 0)
	// query2("SELECT id from role where time_login < '2019-03-25 00:00:00'", func(rows *sql.Rows) {
	// 	var rowData rowID
	// 	rows.Scan(&rowData.ID)
	// 	result2 = append(result2, rowData)
	// })
	// fmt.Println(result2)

	redisTest()
}
