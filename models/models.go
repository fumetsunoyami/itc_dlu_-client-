package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"myapp/database"
	"time"
)

type Student struct {
	ID           int
	StudentCode  string
	Name         string
	DateOfBirth  string
	Gender       string
	PlaceOfBirth string
}

type Exam struct {
	ID            int
	Name          string
	Duration      int
	QuestionSetID int
}

type Question struct {
	ID            int
	Content       string
	Answers       []string
	CorrectAnswer string
}

type Answer struct {
	QuestionID    int
	Selected      string
	CorrectAnswer string
}

func CheckStudent(studentCode string) (*Student, error) {
	var student Student
	query := "SELECT id, student_code, name, date_of_birth, gender, place_of_birth FROM students WHERE student_code = ?"
	err := database.DB.QueryRow(query, studentCode).Scan(&student.ID, &student.StudentCode, &student.Name, &student.DateOfBirth, &student.Gender, &student.PlaceOfBirth)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &student, nil
}

func GetExam(examID int) (*Exam, error) {
	var exam Exam
	query := `
		SELECT e.id, e.name, ec.duration, e.question_set_id 
		FROM exams e 
		JOIN exam_config ec ON e.id = ec.exam_id 
		WHERE e.id = ?`
	err := database.DB.QueryRow(query, examID).Scan(&exam.ID, &exam.Name, &exam.Duration, &exam.QuestionSetID)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &exam, nil
}

func GetQuestions(examID int) ([]Question, error) {
	var questions []Question
	query := `
		SELECT q.id, q.content, qa.answers, qa.correct_answer
		FROM questions q
		JOIN question_answers qa ON q.id = qa.question_id
		JOIN question_set_questions qsq ON q.id = qsq.question_id
		JOIN exams e ON e.question_set_id = qsq.question_set_id
		WHERE e.id = ?`
	rows, err := database.DB.Query(query, examID)
	if err != nil {
		return nil, fmt.Errorf("lỗi truy vấn cơ sở dữ liệu: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var q Question
		var answersJSON string
		err := rows.Scan(&q.ID, &q.Content, &answersJSON, &q.CorrectAnswer)
		if err != nil {
			return nil, fmt.Errorf("lỗi quét dữ liệu: %v", err)
		}
		if err := json.Unmarshal([]byte(answersJSON), &q.Answers); err != nil {
			return nil, fmt.Errorf("lỗi parse JSON answers: %v", err)
		}
		questions = append(questions, q)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("lỗi duyệt rows: %v", err)
	}
	return questions, nil
}

func StartExamStudent(examID, studentID int, duration int) error {
	// Kiểm tra xem bản ghi đã tồn tại chưa
	var existingID int
	var currentStatus string
	err := database.DB.QueryRow(`
		SELECT id, status FROM exam_students 
		WHERE exam_id = ? AND student_id = ?`,
		examID, studentID).Scan(&existingID, &currentStatus)

	if err == sql.ErrNoRows {
		// Nếu chưa tồn tại, thêm mới với trạng thái 'in_progress'
		_, err = database.DB.Exec(`
			INSERT INTO exam_students (exam_id, student_id, group_id, status, test_started_at, remaining_time)
			VALUES (?, ?, 1, 'in_progress', ?, ?)`,
			examID, studentID, time.Now(), duration*60)
		if err != nil {
			return fmt.Errorf("lỗi khi thêm bản ghi mới vào exam_students: %v", err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("lỗi khi kiểm tra bản ghi trong exam_students: %v", err)
	}

	// Nếu đã tồn tại, kiểm tra trạng thái
	if currentStatus == "completed" {
		return fmt.Errorf("kỳ thi đã hoàn thành, không thể bắt đầu lại")
	}

	// Cập nhật trạng thái và thời gian
	_, err = database.DB.Exec(`
		UPDATE exam_students 
		SET status = 'in_progress', test_started_at = ?, remaining_time = ?
		WHERE exam_id = ? AND student_id = ?`,
		time.Now(), duration*60, examID, studentID)
	if err != nil {
		return fmt.Errorf("lỗi khi cập nhật trạng thái trong exam_students: %v", err)
	}
	return nil
}

func UpdateRemainingTime(examID, studentID, remainingTime int) error {
	_, err := database.DB.Exec(`
		UPDATE exam_students 
		SET remaining_time = ?, status = 'in_progress'
		WHERE exam_id = ? AND student_id = ?`,
		remainingTime, examID, studentID)
	if err != nil {
		return fmt.Errorf("lỗi khi cập nhật thời gian còn lại: %v", err)
	}
	return nil
}

func CompleteExamStudent(examID, studentID int) error {
	_, err := database.DB.Exec(`
		UPDATE exam_students 
		SET status = 'completed', remaining_time = 0
		WHERE exam_id = ? AND student_id = ?`,
		examID, studentID)
	if err != nil {
		return fmt.Errorf("lỗi khi cập nhật trạng thái hoàn thành: %v", err)
	}
	return nil
}
