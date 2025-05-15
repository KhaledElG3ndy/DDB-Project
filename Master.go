package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

var clients = make(map[string]net.Conn)
var mu sync.Mutex
var db *sql.DB

type QueryRequest struct {
	Query string `json:"query"`
}

type QueryResponse struct {
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func handleConnection(conn net.Conn) {
	addr := conn.RemoteAddr().String()
	mu.Lock()
	clients[addr] = conn
	mu.Unlock()
	fmt.Println("Client connected:", addr)

	defer func() {
		mu.Lock()
		delete(clients, addr)
		mu.Unlock()
		conn.Close()
		fmt.Println("Client disconnected:", addr)
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		query := strings.TrimSpace(scanner.Text())
		if query == "" {
			continue
		}
		go func(q string, c net.Conn) {
			fmt.Printf("Executing from %s: %s\n", addr, q)
			result, err := executeQuery(q)
			if err != nil {
				c.Write([]byte("ERROR: " + err.Error() + "\n"))
			} else {
				c.Write([]byte("RESULT:\n" + result + "\n"))
			}
		}(query, conn)
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

		var result strings.Builder
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
				raw := values[i]
				if b, ok := raw.([]byte); ok {
					result.WriteString(fmt.Sprintf("%s: %s\t", col, string(b)))
				} else {
					result.WriteString(fmt.Sprintf("%s: %v\t", col, raw))
				}
			}
			result.WriteString("\n")
		}
		return result.String(), nil
	}

	res, err := db.Exec(query)
	if err != nil {
		return "", err
	}
	affected, _ := res.RowsAffected()
	return fmt.Sprintf("Query OK, %d rows affected", affected), nil
}

func handleQueryAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := executeQuery(req.Query)
	resp := QueryResponse{}
	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.Result = result
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	var err error
	dsn := "root:rootroot@tcp(127.0.0.1:3306)/ddbproject?parseTime=true"
	db, err = sql.Open("mysql", dsn)
	if err != nil || db.Ping() != nil {
		fmt.Println("Database connection failed:", err)
		return
	}
	fmt.Println("Connected to MySQL database")

	go func() {
		ln, err := net.Listen("tcp", ":9999")
		if err != nil {
			fmt.Println("TCP server error:", err)
			return
		}
		fmt.Println("TCP server started on port 9999")
		for {
			conn, err := ln.Accept()
			if err == nil {
				go handleConnection(conn)
			}
		}
	}()

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/query", handleQueryAPI)

	fmt.Println("HTTP server started on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
