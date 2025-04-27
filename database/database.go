package database

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func ConnectDB() {
	var err error
	dsn := "root:@tcp(127.0.0.1:3306)/exam_system?charset=utf8mb4&parseTime=True"
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Lỗi kết nối MySQL: %v", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatalf("Không thể ping MySQL: %v", err)
	}

	log.Println("Kết nối MySQL thành công!")
}
