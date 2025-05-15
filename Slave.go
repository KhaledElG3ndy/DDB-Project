package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"net"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var backup bool
func main() {
	var err error
	conn, err := net.Dial("tcp", "127.0.0.1:9999") 
	if err != nil {
		fmt.Println("Failed to connect to server:", err)
		return
	}
	defer conn.Close()
	masterConf := "root:Khaled@l3153928@tcp(127.0.0.1:3306)/ddbproject" 
	slaveConf := "root:Khaled@l3153928@tcp(127.0.0.1:3306)/backup"
	db, err = sql.Open("mysql", masterConf)
	if err != nil || db.Ping() != nil {
		fmt.Println("Failed to connect to master's DB, switching to local backup...")
		backup = true
		db, err = sql.Open("mysql", slaveConf)
		if err != nil || db.Ping() != nil {
			fmt.Println("Failed to connect to local backup DB as well. Exiting.")
			return
		}
		fmt.Println("Using local backup database")
	} else {
		fmt.Println("Connected to master's database")
		backup = false
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter SQL query (or exit to quit): ")
		query, _ := reader.ReadString('\n')
		query = strings.TrimSpace(query)
		if query == "exit" {
			break
		}
		if query == "" {
			continue
		}
		res, err := executeQuery(query)
		if err != nil {
			fmt.Println("Error executing query:", err)
		} else {
			fmt.Println("Result:", res)
		}
	}
}

func executeQuery(query string) (string, error) {
	queryLower := strings.ToLower(query)
	if strings.HasPrefix(queryLower, "select") {
		rows, err := db.Query(query)
		if err != nil {
			return "", err
		}
		defer rows.Close()

		cols, err := rows.Columns()
		if err != nil {
			return "", err
		}

		result := ""
		for rows.Next() {
			values := make([]interface{}, len(cols))
			valuePtrs := make([]interface{}, len(cols))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				return "", err
			}

			for i, col := range cols {
				result += fmt.Sprintf("%s: %v\t", col, values[i])
			}
			result += "\n"
		}
		return result, nil
	} else {
		res, err := db.Exec(query)
		if err != nil {
			return "", err
		}
		affected, _ := res.RowsAffected()

		if !backup {
			slaveConf := "root:Khaled@l3153928@tcp(127.0.0.1:3306)/backup"
			backupDB, err := sql.Open("mysql", slaveConf)
			if err == nil && backupDB.Ping() == nil {
				defer backupDB.Close()
				_, _ = backupDB.Exec(query) 
			}
		}
		return fmt.Sprintf("Query OK, %d rows affected", affected), nil
	}
}
