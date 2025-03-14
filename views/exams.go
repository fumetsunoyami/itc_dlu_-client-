package views

import (
	"fmt"
	"myapp/controllers"
	"myapp/models"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func ShowExam(myApp fyne.App, student models.Student, examID int) {
	myWindow := myApp.NewWindow("Hệ thống Trắc nghiệm")
	myWindow.Resize(fyne.NewSize(900, 600))

	// Lấy thông tin kỳ thi
	exam, err := controllers.GetExamInfo(examID)
	if err != nil || exam == nil {
		myWindow.SetContent(widget.NewLabel("Lỗi tải thông tin kỳ thi"))
		return
	}

	// Khởi tạo trạng thái kỳ thi
	err = controllers.StartExam(examID, student.ID, exam.Duration)
	if err != nil {
		myWindow.SetContent(widget.NewLabel("Lỗi khởi tạo kỳ thi"))
		return
	}

	// MSSV (góc trên bên trái)
	mssvLabel := widget.NewLabel("Mã số sinh viên: " + student.StudentCode)
	mssvLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Đồng hồ đếm ngược (góc trên bên phải)
	timeLeft := exam.Duration * 60
	timer := widget.NewLabel(formatTime(timeLeft))
	timer.TextStyle = fyne.TextStyle{Bold: true}

	// Lấy danh sách câu hỏi
	questions, err := controllers.GetExamQuestions(exam.ID)
	if err != nil {
		myWindow.SetContent(widget.NewLabel("Lỗi tải câu hỏi"))
		return
	}

	// Lưới câu hỏi (bên phải)
	numCols := 5
	questionGrid := container.NewGridWithColumns(numCols)
	answers := make(map[int]string)
	answeredCount := 0
	answeredLabel := widget.NewLabel(fmt.Sprintf("Số câu: %d/%d", answeredCount, len(questions)))

	// Phần câu hỏi (bên trái, dạng cuộn vô hạn)
	questionWidgets := make([]fyne.CanvasObject, len(questions))
	questionContainer := container.NewVScroll(container.NewVBox()) // Khai báo trước để dùng

	// Biến để theo dõi các nút
	var buttons []*widget.Button

	for i := 0; i < len(questions); i++ {
		i := i
		// Tạo nút cho lưới câu hỏi
		btn := widget.NewButton(fmt.Sprintf("%d", i+1), func() {
			// Cuộn thủ công bằng cách điều chỉnh Offset
			estimatedHeightPerQuestion := float32(150)
			totalHeight := float32(len(questions)) * estimatedHeightPerQuestion
			targetOffset := float32(i) * estimatedHeightPerQuestion / totalHeight
			questionContainer.Offset.Y = targetOffset
			questionContainer.Refresh()

			// Làm nổi bật nút được chọn
			for _, b := range buttons {
				b.Importance = widget.MediumImportance
			}
			buttons[i].Importance = widget.HighImportance
			questionGrid.Refresh()
		})
		buttons = append(buttons, btn)
		questionGrid.Add(btn)

		// Tạo widget cho câu hỏi
		questionLabel := widget.NewLabel(fmt.Sprintf("Câu %d: %s", i+1, questions[i].Content))
		questionLabel.TextStyle = fyne.TextStyle{Bold: true}
		options := widget.NewRadioGroup([]string{
			questions[i].OptionA, questions[i].OptionB,
			questions[i].OptionC, questions[i].OptionD,
		}, func(selected string) {
			if answers[questions[i].ID] == "" && selected != "" {
				answeredCount++
				answeredLabel.SetText(fmt.Sprintf("Số câu: %d/%d", answeredCount, len(questions)))
			}
			answers[questions[i].ID] = selected
		})
		options.Horizontal = false
		questionWidgets[i] = container.NewVBox(
			questionLabel,
			options,
			widget.NewSeparator(),
		)
	}

	// Thêm các widget câu hỏi vào container
	questionContainer.Content = container.NewVBox(questionWidgets...)
	questionContainer.SetMinSize(fyne.NewSize(600, 500))

	// Đếm ngược thời gian
	go func() {
		for timeLeft > 0 {
			time.Sleep(1 * time.Second)
			timeLeft--
			timer.SetText(formatTime(timeLeft))
			if timeLeft%10 == 0 {
				err := controllers.UpdateRemainingTime(examID, student.ID, timeLeft)
				if err != nil {
					// Có thể thêm log lỗi nếu cần
				}
			}
		}
		finalScore, err := controllers.SubmitExam(examID, student.ID, questions, answers)
		if err != nil {
			myWindow.SetContent(widget.NewLabel("Lỗi khi nộp bài"))
			return
		}
		myWindow.SetContent(widget.NewLabel(fmt.Sprintf("Bài thi đã được nộp! Điểm: %.2f", finalScore)))
	}()

	// Nút nộp bài (dưới cùng bên phải)
	submitButton := widget.NewButton("Nộp bài", func() {
		finalScore, err := controllers.SubmitExam(examID, student.ID, questions, answers)
		if err != nil {
			myWindow.SetContent(widget.NewLabel("Lỗi khi nộp bài"))
			return
		}
		myWindow.SetContent(widget.NewLabel(fmt.Sprintf("Bài thi đã được nộp! Điểm: %.2f", finalScore)))
	})
	submitButton.Importance = widget.HighImportance

	// Bố cục lưới câu hỏi và số câu làm được
	rightPanel := container.NewVBox(
		answeredLabel,
		questionGrid,
		submitButton,
	)
	rightContainer := container.NewScroll(rightPanel)
	rightContainer.SetMinSize(fyne.NewSize(200, 500))

	// Thanh trên cùng
	topBar := container.NewHBox(
		mssvLabel,
		layout.NewSpacer(),
		timer,
	)

	// Bố cục chính
	split := container.NewHSplit(questionContainer, rightContainer)
	split.SetOffset(0.75) // Phần câu hỏi chiếm 75%, lưới câu hỏi chiếm 25%

	mainContainer := container.NewBorder(
		container.NewVBox(topBar, canvas.NewLine(theme.ForegroundColor())),
		nil, nil, nil, split,
	)

	myWindow.SetContent(mainContainer)
	myWindow.Show()
}

func formatTime(seconds int) string {
	minutes := seconds / 60
	secs := seconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, secs)
}
