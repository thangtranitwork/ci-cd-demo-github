package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitMySQL() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		getEnv("MYSQL_USER", "root"),
		getEnv("MYSQL_PASSWORD", "rootsecret"),
		getEnv("MYSQL_HOST", "localhost"),
		getEnv("MYSQL_PORT", "3306"),
		getEnv("MYSQL_DATABASE", "feature_flags"),
	)

	var err error
	for i := 0; i < 10; i++ {
		DB, err = sql.Open("mysql", dsn)
		if err == nil {
			err = DB.Ping()
		}
		if err == nil {
			break
		}
		log.Printf("⏳ MySQL chưa sẵn sàng, thử lại sau 3s... (%d/10)", i+1)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatalf("❌ Không thể kết nối MySQL: %v", err)
	}

	log.Println("✅ Kết nối MySQL thành công")
	runMigrations()
}

func runMigrations() {
	migration := `
	CREATE TABLE IF NOT EXISTS features (
		id          INT AUTO_INCREMENT PRIMARY KEY,
		name        VARCHAR(100) NOT NULL UNIQUE,
		description TEXT,
		enabled     BOOLEAN NOT NULL DEFAULT TRUE,
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS feature_routes (
		id          INT AUTO_INCREMENT PRIMARY KEY,
		feature_id  INT NOT NULL,
		method      VARCHAR(10) NOT NULL,
		path        VARCHAR(255) NOT NULL,
		enabled     BOOLEAN NOT NULL DEFAULT TRUE,
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (feature_id) REFERENCES features(id) ON DELETE CASCADE,
		UNIQUE KEY uq_method_path (method, path)
	);
	`

	for _, stmt := range splitSQL(migration) {
		if stmt == "" {
			continue
		}
		if _, err := DB.Exec(stmt); err != nil {
			log.Fatalf("❌ Migration thất bại: %v\nSQL: %s", err, stmt)
		}
	}
	log.Println("✅ Migration MySQL hoàn tất")

	seedData()
}

func seedData() {
	// Kiểm tra xem đã có dữ liệu chưa
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM features").Scan(&count)
	if count > 0 {
		return
	}

	log.Println("🌱 Seeding dữ liệu ban đầu...")

	seeds := []struct {
		feature string
		desc    string
		routes  [][3]string // method, path, enabled
	}{
		{
			feature: "user-management",
			desc:    "Quản lý người dùng",
			routes: [][3]string{
				{"GET", "/api/users", "1"},
				{"POST", "/api/users", "1"},
				{"DELETE", "/api/users/{id}", "1"},
			},
		},
		{
			feature: "product-catalog",
			desc:    "Danh mục sản phẩm",
			routes: [][3]string{
				{"GET", "/api/products", "1"},
				{"POST", "/api/products", "1"},
				{"GET", "/api/products/{id}", "1"},
				{"PUT", "/api/products/{id}", "0"},
			},
		},
		{
			feature: "beta-dashboard",
			desc:    "Dashboard beta (đang thử nghiệm)",
			routes: [][3]string{
				{"GET", "/api/dashboard", "0"},
				{"GET", "/api/dashboard/stats", "0"},
			},
		},
	}

	for _, s := range seeds {
		res, err := DB.Exec(
			"INSERT INTO features (name, description, enabled) VALUES (?, ?, ?)",
			s.feature, s.desc, true,
		)
		if err != nil {
			log.Printf("⚠️ Seed feature '%s': %v", s.feature, err)
			continue
		}
		fid, _ := res.LastInsertId()
		for _, r := range s.routes {
			enabled := r[2] == "1"
			DB.Exec(
				"INSERT INTO feature_routes (feature_id, method, path, enabled) VALUES (?, ?, ?, ?)",
				fid, r[0], r[1], enabled,
			)
		}
	}
	log.Println("✅ Seed dữ liệu hoàn tất")
}

func splitSQL(sql string) []string {
	var stmts []string
	var current string
	for _, ch := range sql {
		if ch == ';' {
			stmts = append(stmts, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	return stmts
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
