package models

import (
	"database/sql"
	"myapp/database"
	"time"
)

type Student struct {
	ID          int
	StudentCode string
	Name        string
	Class       string
	Faculty     string
	CCCD        string
}

type Exam struct {
	ID            int
	Name          string
	Duration      int // Thời gian làm bài (phút)
	QuestionSetID int
}

type Question struct {
	ID            int
	Content       string
	OptionA       string
	OptionB       string
	OptionC       string
	OptionD       string
	CorrectAnswer string
}

func CheckStudent(studentCode string) (*Student, error) {
	var student Student
	query := "SELECT id, student_code, name, class, faculty, cccd FROM students WHERE student_code = ?"
	err := database.DB.QueryRow(query, studentCode).Scan(&student.ID, &student.StudentCode, &student.Name, &student.Class, &student.Faculty, &student.CCCD)
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
		SELECT q.id, q.content, q.option_A, q.option_B, q.option_C, q.option_D, q.correct_answer
		FROM questions q
		JOIN question_set_questions qsq ON q.id = qsq.question_id
		JOIN exams e ON e.question_set_id = qsq.question_set_id
		WHERE e.id = ?`
	rows, err := database.DB.Query(query, examID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var q Question
		err := rows.Scan(&q.ID, &q.Content, &q.OptionA, &q.OptionB, &q.OptionC, &q.OptionD, &q.CorrectAnswer)
		if err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	return questions, nil
}

func StartExamStudent(examID, studentID int, duration int) error {
	_, err := database.DB.Exec(`
		INSERT INTO exam_students (exam_id, student_id, group_id, status, test_started_at, remaining_time)
		VALUES (?, ?, 1, 'in_progress', ?, ?)`,
		examID, studentID, time.Now(), duration*60)
	return err
}

func UpdateRemainingTime(examID, studentID, remainingTime int) error {
	_, err := database.DB.Exec(`
		UPDATE exam_students 
		SET remaining_time = ?, status = 'in_progress'
		WHERE exam_id = ? AND student_id = ?`,
		remainingTime, examID, studentID)
	return err
}

func CompleteExamStudent(examID, studentID int) error {
	_, err := database.DB.Exec(`
		UPDATE exam_students 
		SET status = 'completed', remaining_time = 0
		WHERE exam_id = ? AND student_id = ?`,
		examID, studentID)
	return err
}
