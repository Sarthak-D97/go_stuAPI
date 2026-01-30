package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/Sarthak-D97/go_stuAPI/internal/config"
	"github.com/Sarthak-D97/go_stuAPI/internal/types"
	_ "github.com/mattn/go-sqlite3"
)

type Sqlite struct {
	Db *sql.DB
}

func New(cfg *config.Config) (*Sqlite, error) {
	db, err := sql.Open("sqlite3", cfg.StoragePath)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA busy_timeout = 5000;"); err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS students(
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        email TEXT NOT NULL UNIQUE,
        age INTEGER NOT NULL
    );`)

	if err != nil {
		return nil, err
	}

	return &Sqlite{Db: db}, nil
}

func (s *Sqlite) CreateStudent(name string, email string, age int) (int64, error) {
	query := "INSERT INTO students (name, email, age) VALUES (?, ?, ?)"
	result, err := s.Db.Exec(query, name, email, age)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (s *Sqlite) GetStudentById(id int64) (*types.Student, error) {
	query := "SELECT id, name, email, age FROM students WHERE id = ? LIMIT 1"
	var student types.Student
	err := s.Db.QueryRow(query, id).Scan(&student.ID, &student.Name, &student.Email, &student.Age)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("student not found")
		}
		return nil, err
	}
	return &student, nil
}

func (s *Sqlite) GetAllStudents() ([]types.Student, error) {
	query := "SELECT id, name, email, age FROM students"
	rows, err := s.Db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []types.Student
	for rows.Next() {
		var student types.Student
		if err := rows.Scan(&student.ID, &student.Name, &student.Email, &student.Age); err != nil {
			return nil, err
		}
		students = append(students, student)
	}
	return students, rows.Err()
}

func (s *Sqlite) UpdateStudent(id int64, name string, email string, age int) error {
	query := "UPDATE students SET name = ?, email = ?, age = ? WHERE id = ?"
	result, err := s.Db.Exec(query, name, email, age, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err == nil && rows == 0 {
		return fmt.Errorf("not found")
	}
	return err
}

func (s *Sqlite) DeleteStudent(id int64) error {
	query := "DELETE FROM students WHERE id = ?"
	result, err := s.Db.Exec(query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err == nil && rows == 0 {
		return fmt.Errorf("not found")
	}
	return err
}
