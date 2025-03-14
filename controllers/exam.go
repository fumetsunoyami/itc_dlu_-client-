package controllers

import (
	"myapp/database"
	"myapp/models"
)

var DB = database.DB // Biến DB để sử dụng trong controller

// GetExamQuestions lấy danh sách câu hỏi cho một kỳ thi
func GetExamQuestions(examID int) ([]models.Question, error) {
	questions, err := models.GetQuestions(examID)
	if err != nil {
		return nil, err
	}
	return questions, nil
}

// GetExamInfo lấy thông tin kỳ thi
func GetExamInfo(examID int) (*models.Exam, error) {
	exam, err := models.GetExam(examID)
	if err != nil {
		return nil, err
	}
	if exam == nil {
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
	score := 0.0
	for _, q := range questions {
		if answers[q.ID] == q.CorrectAnswer {
			score++
		}
	}
	finalScore := (score / float64(len(questions))) * 10

	// Lưu kết quả vào exam_results
	_, err := DB.Exec(`
		INSERT INTO exam_results (exam_id, student_id, score) 
		VALUES (?, ?, ?)`, examID, studentID, finalScore)
	if err != nil {
		return 0, err
	}

	// Cập nhật trạng thái hoàn thành
	err = models.CompleteExamStudent(examID, studentID)
	if err != nil {
		return 0, err
	}

	return finalScore, nil
}
