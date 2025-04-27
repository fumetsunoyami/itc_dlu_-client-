package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"myapp/database"
	"myapp/models"
)

var DB = database.DB // Biến DB để sử dụng trong controller

// GetExamQuestions lấy danh sách câu hỏi cho một kỳ thi
func GetExamQuestions(examID int) ([]models.Question, error) {
	// Kiểm tra kết nối cơ sở dữ liệu
	if database.DB == nil {
		return nil, fmt.Errorf("CSDL chưa được khởi tạo")
	}

	rows, err := database.DB.Query(`
		SELECT q.id, q.content, qa.answers, qa.correct_answer
		FROM questions q
		JOIN question_answers qa ON q.id = qa.question_id
		JOIN question_set_questions qsq ON q.id = qsq.question_id
		JOIN exams e ON e.question_set_id = qsq.question_set_id
		WHERE e.id = ?`, examID)
	if err != nil {
		log.Printf("Lỗi truy vấn cơ sở dữ liệu trong GetExamQuestions: %v", err)
		return nil, fmt.Errorf("lỗi truy vấn cơ sở dữ liệu: %v", err)
	}
	defer rows.Close()

	var questions []models.Question
	for rows.Next() {
		var q models.Question
		var answersJSON string
		err := rows.Scan(&q.ID, &q.Content, &answersJSON, &q.CorrectAnswer)
		if err != nil {
			log.Printf("Lỗi quét dữ liệu trong GetExamQuestions: %v", err)
			return nil, fmt.Errorf("lỗi quét dữ liệu: %v", err)
		}
		// Phân tích JSON từ cột answers
		if err := json.Unmarshal([]byte(answersJSON), &q.Answers); err != nil {
			log.Printf("Lỗi parse JSON answers cho question_id %d: %v", q.ID, err)
			return nil, fmt.Errorf("lỗi parse JSON answers: %v", err)
		}
		questions = append(questions, q)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Lỗi duyệt rows trong GetExamQuestions: %v", err)
		return nil, fmt.Errorf("lỗi duyệt rows: %v", err)
	}

	if len(questions) == 0 {
		log.Printf("Không tìm thấy câu hỏi nào cho exam_id %d", examID)
		return nil, fmt.Errorf("không tìm thấy câu hỏi nào cho kỳ thi này")
	}

	return questions, nil
}

// GetExamInfo lấy thông tin kỳ thi
func GetExamInfo(examID int) (*models.Exam, error) {
	exam, err := models.GetExam(examID)
	if err != nil {
		log.Printf("Lỗi lấy thông tin kỳ thi exam_id %d: %v", examID, err)
		return nil, err
	}
	if exam == nil {
		log.Printf("Không tìm thấy kỳ thi với exam_id %d", examID)
		return nil, nil
	}
	return exam, nil
}

// StartExam khởi tạo trạng thái kỳ thi cho sinh viên
func StartExam(examID, studentID, duration int) error {
	return models.StartExamStudent(examID, studentID, duration)
}

// UpdateRemainingTime cập nhật thời gian còn lại
func UpdateRemainingTime(examID, studentID, remainingTime int) error {
	return models.UpdateRemainingTime(examID, studentID, remainingTime)
}

// CompleteExam hoàn thành kỳ thi
func CompleteExam(examID, studentID int) error {
	return models.CompleteExamStudent(examID, studentID)
}

// SubmitExam tính điểm và lưu kết quả
func SubmitExam(examID, studentID int, questions []models.Question, answers map[int]string) (float64, error) {
	// Kiểm tra xem kết nối cơ sở dữ liệu có sẵn không
	if database.DB == nil {
		return 0, fmt.Errorf("CSDL chưa được khởi tạo")
	}

	// Bắt đầu giao dịch
	tx, err := database.DB.Begin()
	if err != nil {
		log.Printf("Lỗi bắt đầu giao dịch trong SubmitExam: %v", err)
		return 0, fmt.Errorf("lỗi bắt đầu giao dịch: %v", err)
	}
	defer tx.Rollback() // Rollback nếu có lỗi

	// Tính điểm
	correctCount := 0
	for _, question := range questions {
		if selected, exists := answers[question.ID]; exists && selected == question.CorrectAnswer {
			correctCount++
		}
	}
	finalScore := float64(correctCount) / float64(len(questions)) * 10

	// Lưu kết quả vào bảng exam_results
	_, err = tx.Exec(
		"INSERT INTO exam_results (exam_id, student_id, score) VALUES (?, ?, ?)",
		examID, studentID, finalScore,
	)
	if err != nil {
		log.Printf("Lỗi lưu kết quả kỳ thi trong SubmitExam: %v", err)
		return 0, fmt.Errorf("lỗi lưu kết quả kỳ thi: %v", err)
	}

	// Cập nhật trạng thái trong exam_students thành 'completed'
	_, err = tx.Exec(`
		UPDATE exam_students 
		SET status = 'completed', remaining_time = 0
		WHERE exam_id = ? AND student_id = ?`,
		examID, studentID)
	if err != nil {
		log.Printf("Lỗi cập nhật trạng thái exam_students trong SubmitExam: %v", err)
		return 0, fmt.Errorf("lỗi cập nhật trạng thái exam_students: %v", err)
	}

	// Commit giao dịch
	if err := tx.Commit(); err != nil {
		log.Printf("Lỗi commit giao dịch trong SubmitExam: %v", err)
		return 0, fmt.Errorf("lỗi commit giao dịch: %v", err)
	}

	log.Printf("Nộp bài thành công cho sinh viên %d, kỳ thi %d, điểm: %.2f", studentID, examID, finalScore)
	return finalScore, nil
}
