package views

import (
	"fmt"
	"myapp/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func ShowResult(myApp fyne.App, student models.Student, answers []models.Answer, correctCount, totalQuestions int) {
	// Tạo cửa sổ mới
	completeWindow := myApp.NewWindow("Kết thúc bài thi")
	completeWindow.Resize(fyne.NewSize(600, 400))

	// Tính điểm
	finalScore := float64(correctCount) / float64(totalQuestions) * 10

	// Tạo container để chứa nội dung
	resultContainer := container.NewVBox()

	// Hiển thị thông tin sinh viên và điểm
	resultContainer.Add(widget.NewLabel(fmt.Sprintf("Mã số sinh viên: %s", student.StudentCode)))
	resultContainer.Add(widget.NewLabel(fmt.Sprintf("Tên thí sinh: %s", student.Name)))
	resultContainer.Add(widget.NewLabel(fmt.Sprintf("Điểm: %.2f/10", finalScore)))
	resultContainer.Add(widget.NewLabel(fmt.Sprintf("Số câu trả lời đúng: %d/%d", correctCount, totalQuestions)))

	// Hiển thị tiêu đề phần câu trả lời
	resultContainer.Add(widget.NewLabel("Chi tiết câu trả lời:"))

	// Hiển thị từng câu hỏi
	for _, answer := range answers {
		questionInfo := fmt.Sprintf("Câu %d: Bạn chọn '%s' - Đáp án đúng: '%s'",
			answer.QuestionID, answer.Selected, answer.CorrectAnswer)
		resultContainer.Add(widget.NewLabel(questionInfo))
	}

	// Thêm thanh cuộn nếu nội dung dài
	scroll := container.NewVScroll(resultContainer)
	scroll.SetMinSize(fyne.NewSize(580, 350))

	// Nút quay lại
	backButton := widget.NewButton("Quay lại", func() {
		completeWindow.Close()
	})

	// Sắp xếp bố cục
	content := container.NewBorder(nil, backButton, nil, nil, scroll)
	completeWindow.SetContent(content)
	completeWindow.Show()
}
