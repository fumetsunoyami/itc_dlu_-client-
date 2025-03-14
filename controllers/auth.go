package controllers

import (
	"myapp/models"
)

// LoginStudent kiểm tra đăng nhập sinh viên
func LoginStudent(studentCode string) (*models.Student, error) {
	student, err := models.CheckStudent(studentCode)
	if err != nil {
		return nil, err
	}
	if student == nil {
		return nil, nil // Không tìm thấy sinh viên
	}
	return student, nil // Đăng nhập thành công
}
