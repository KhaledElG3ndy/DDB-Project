# 📡 Distributed SQL System with Master-Slave Architecture

This project is a distributed SQL execution system written in Go. It features a **master-slave database architecture** for failover support, **TCP client-server communication**, and a **web-based GUI** for interacting with the system.

---

## 🚀 Features

- ✅ **Master-Slave Failover**  
  Automatically connects to a local backup (slave) database if the master DB is unreachable.

- 🔌 **TCP Client-Server Communication**  
  Supports multiple concurrent clients sending SQL queries over TCP.

- 🌐 **Web-Based Query Interface**  
  Lightweight GUI served via HTTP at `localhost:8080`, allowing users to submit SQL queries through a browser.

- 📜 **Query Execution Engine**  
  Supports `SELECT`, `INSERT`, `UPDATE`, and more — results are returned in formatted output.

- 🔁 **Write Replication**  
  All write queries (non-SELECT) on the master are automatically executed on the slave if available.

- 🧵 **Concurrent Connections**  
  Built using Go routines and mutexes to handle multiple connections safely.

---

## 🗂️ Project Structure

├── main.go # Master server (TCP + HTTP + DB executor)
├── client.go # Client application with automatic DB fallback
├── static/
│ └── index.html # Web GUI for query execution


---

## 💻 How to Run

### 1. Start the MySQL databases

Ensure you have:
- A **master database** (`ddbproject`)
- A **backup database** (`backup`)

Update credentials in the Go code as needed.

### 2. Run the Master Server

```bash
go run main.go
Starts a TCP server on port 9999

Starts an HTTP server with GUI on localhost:8080
3. Run a Client
bash
Copy
Edit
go run client.go
Connects to the master via TCP

Accepts SQL queries from terminal

Falls back to slave database if needed

4. Use the Web GUI
Visit: http://localhost:8080

Submit SQL queries and view results directly from your browser.![image](https://github.com/user-attachments/assets/eecf6855-1e80-4c24-8b2c-8b6ed05019f7)
