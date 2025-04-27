package views

import (
	"fmt"
	"image/color"
	"log"
	"myapp/controllers"
	"myapp/database"
	"myapp/models"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func ShowExam(myApp fyne.App, student models.Student, examID int) {
	myWindow := myApp.NewWindow("Hệ thống Trắc nghiệm")
	myWindow.SetFullScreen(true)

	// Lấy thông tin kỳ thi
	exam, err := controllers.GetExamInfo(examID)
	if err != nil || exam == nil {
		log.Printf("Lỗi tải thông tin kỳ thi exam_id %d: %v", examID, err)
		myWindow.SetContent(widget.NewLabel("Lỗi tải thông tin kỳ thi"))
		myWindow.Show()
		return
	}

	// Lấy exam_student_id từ bảng exam_students
	var examStudentID int
	var status string
	var remainingTime int
	err = database.DB.QueryRow(`
		SELECT id, status, remaining_time 
		FROM exam_students 
		WHERE exam_id = ? AND student_id = ?`,
		examID, student.ID).Scan(&examStudentID, &status, &remainingTime)
	if err != nil {
		log.Printf("Lỗi kiểm tra trạng thái exam_students cho exam_id %d, student_id %d: %v", examID, student.ID, err)
		myWindow.SetContent(widget.NewLabel("Lỗi kiểm tra trạng thái kỳ thi"))
		myWindow.Show()
		return
	}

	// Nếu trạng thái là 'completed', không cho phép làm bài
	if status == "completed" {
		log.Println("Kỳ thi đã hoàn thành, không thể làm bài")
		// Hiển thị giao diện chỉ đọc
		questions, err := controllers.GetExamQuestions(exam.ID)
		if err != nil {
			log.Printf("Lỗi tải câu hỏi cho exam_id %d: %v", exam.ID, err)
			myWindow.SetContent(widget.NewLabel("Lỗi tải câu hỏi: " + err.Error()))
			myWindow.Show()
			return
		}

		answers := make(map[int]string)
		answeredStatus := make(map[int]bool)
		questionWidgets := make([]fyne.CanvasObject, len(questions))

		// Lấy câu trả lời đã lưu từ exam_student_answers
		rows, err := database.DB.Query(`
			SELECT question_id, selected_answer 
			FROM exam_student_answers 
			WHERE exam_student_id = ?`, examStudentID)
		if err != nil {
			log.Printf("Lỗi lấy câu trả lời từ exam_student_answers: %v", err)
		} else {
			defer rows.Close()
			for rows.Next() {
				var questionID int
				var selectedAnswer string
				if err := rows.Scan(&questionID, &selectedAnswer); err != nil {
					log.Printf("Lỗi scan câu trả lời từ exam_student_answers: %v", err)
					continue
				}
				answers[questionID] = selectedAnswer
				answeredStatus[questionID] = true
			}
		}

		// Tạo giao diện chỉ đọc
		for i := 0; i < len(questions); i++ {
			questionLabel := widget.NewLabel(fmt.Sprintf("Câu %d: %s", i+1, questions[i].Content))
			questionLabel.TextStyle = fyne.TextStyle{Bold: true}
			var answerLabelText string
			if selected, exists := answers[questions[i].ID]; exists {
				answerLabelText = fmt.Sprintf("Đáp án đã chọn: %s", selected)
			} else {
				answerLabelText = "Chưa trả lời"
			}
			answerLabel := widget.NewLabel(answerLabelText)
			questionWidgets[i] = container.NewVBox(
				questionLabel,
				answerLabel,
				widget.NewSeparator(),
				layout.NewSpacer(),
			)
		}

		questionContainer := container.NewVScroll(container.NewVBox(questionWidgets...))
		questionContainer.SetMinSize(fyne.NewSize(600, 500))

		// Lấy điểm số từ exam_results
		var score float32
		err = database.DB.QueryRow(`
			SELECT score 
			FROM exam_results 
			WHERE exam_id = ? AND student_id = ?`,
			examID, student.ID).Scan(&score)
		if err != nil {
			log.Printf("Lỗi lấy điểm số từ exam_results: %v", err)
			score = 0
		}

		resultLabel := widget.NewLabel(fmt.Sprintf("Kỳ thi đã hoàn thành.\nĐiểm số: %.1f", score))
		resultLabel.TextStyle = fyne.TextStyle{Bold: true}

		backButton := widget.NewButton("Quay lại", func() {
			myWindow.Close()
			ShowStudentInfoWindow(myApp, student)
		})

		content := container.NewVBox(
			resultLabel,
			questionContainer,
			backButton,
		)

		myWindow.SetContent(content)
		myWindow.Show()
		return
	}

	// Nếu trạng thái là 'pending', không cho phép làm bài
	if status == "pending" {
		log.Println("Kỳ thi đang ở trạng thái 'pending', chưa thể làm bài")
		myWindow.SetContent(widget.NewLabel("Kỳ thi chưa bắt đầu, vui lòng chờ!"))
		myWindow.Show()
		return
	}

	// Nếu trạng thái là 'in_progress', cho phép làm bài
	if status == "in_progress" {
		log.Println("Kỳ thi đang ở trạng thái 'in_progress', cho phép làm bài")
	} else {
		// Trường hợp không mong đợi
		log.Printf("Trạng thái kỳ thi không hợp lệ: %s", status)
		myWindow.SetContent(widget.NewLabel("Trạng thái kỳ thi không hợp lệ"))
		myWindow.Show()
		return
	}

	// Lấy danh sách câu hỏi
	questions, err := controllers.GetExamQuestions(exam.ID)
	if err != nil {
		log.Printf("Lỗi tải câu hỏi cho exam_id %d: %v", exam.ID, err)
		myWindow.SetContent(widget.NewLabel("Lỗi tải câu hỏi: " + err.Error()))
		myWindow.Show()
		return
	}

	// Thiết lập giao diện
	mssvLabel := widget.NewLabel("Mã số sinh viên: " + student.StudentCode)
	mssvLabel.TextStyle = fyne.TextStyle{Bold: true}

	timeLeft := remainingTime
	timer := widget.NewLabel(formatTime(timeLeft))
	timer.TextStyle = fyne.TextStyle{Bold: true}

	numCols := 5
	questionGrid := container.NewGridWithColumns(numCols)
	answers := make(map[int]string)
	answeredCount := 0
	answeredLabel := widget.NewLabel(fmt.Sprintf("Số câu: %d/%d", answeredCount, len(questions)))

	answeredStatus := make(map[int]bool)

	questionWidgets := make([]fyne.CanvasObject, len(questions))

	questionContainer := container.NewVScroll(container.NewVBox())

	// Lưu trữ các nút và hình chữ nhật nền để cập nhật màu
	var buttons []*widget.Button
	var backgrounds []*canvas.Rectangle

	// Khôi phục câu trả lời từ bảng exam_student_answers
	rows, err := database.DB.Query(`
		SELECT question_id, selected_answer 
		FROM exam_student_answers 
		WHERE exam_student_id = ?`, examStudentID)
	if err != nil {
		log.Printf("Lỗi lấy câu trả lời từ exam_student_answers: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var questionID int
			var selectedAnswer string
			if err := rows.Scan(&questionID, &selectedAnswer); err != nil {
				log.Printf("Lỗi scan câu trả lời từ exam_student_answers: %v", err)
				continue
			}
			answers[questionID] = selectedAnswer
			answeredStatus[questionID] = true
			answeredCount++
		}
	}

	// Tạo giao diện câu hỏi
	for i := 0; i < len(questions); i++ {
		i := i
		// Tạo hình chữ nhật làm nền
		bg := canvas.NewRectangle(theme.BackgroundColor())
		bg.SetMinSize(fyne.NewSize(40, 40))

		// Tạo nút
		btn := widget.NewButton(fmt.Sprintf("%d", i+1), func() {
			estimatedHeightPerQuestion := float32(150)
			totalHeight := float32(len(questions)) * estimatedHeightPerQuestion
			targetOffset := float32(i) * estimatedHeightPerQuestion / totalHeight
			questionContainer.Offset.Y = targetOffset
			questionContainer.Refresh()
			for _, b := range buttons {
				b.Importance = widget.MediumImportance
			}
			buttons[i].Importance = widget.HighImportance
			questionGrid.Refresh()
		})

		// Đặt nút lên trên hình chữ nhật nền
		cell := container.NewStack(bg, btn)
		buttons = append(buttons, btn)
		backgrounds = append(backgrounds, bg)
		questionGrid.Add(cell)

		// Khôi phục trạng thái màu nền
		if answeredStatus[questions[i].ID] {
			backgrounds[i].FillColor = color.RGBA{R: 0, G: 255, B: 0, A: 255} // Màu xanh lá
			backgrounds[i].Refresh()
		}

		questionLabel := widget.NewLabel(fmt.Sprintf("Câu %d: %s", i+1, questions[i].Content))
		questionLabel.TextStyle = fyne.TextStyle{Bold: true}
		options := widget.NewRadioGroup(questions[i].Answers, func(selected string) {
			oldAnswer, exists := answers[questions[i].ID]
			if !exists {
				oldAnswer = ""
			}
			if selected != "" {
				answers[questions[i].ID] = selected
			} else {
				delete(answers, questions[i].ID)
			}

			if oldAnswer == "" && selected != "" {
				answeredCount++
				answeredStatus[questions[i].ID] = true
				// Cập nhật màu xanh khi trả lời
				backgrounds[i].FillColor = color.RGBA{R: 0, G: 255, B: 0, A: 255} // Màu xanh lá
			} else if oldAnswer != "" && selected == "" {
				answeredCount--
				answeredStatus[questions[i].ID] = false
				// Quay lại màu mặc định khi bỏ chọn
				backgrounds[i].FillColor = theme.BackgroundColor()
			}
			answeredLabel.SetText(fmt.Sprintf("Số câu: %d/%d", answeredCount, len(questions)))
			backgrounds[i].Refresh()
			questionGrid.Refresh()
		})

		// Khôi phục đáp án đã chọn từ bảng exam_student_answers
		if selected, exists := answers[questions[i].ID]; exists {
			options.SetSelected(selected)
		}

		options.Horizontal = true

		questionWidgets[i] = container.NewVBox(
			questionLabel,
			options,
			widget.NewSeparator(),
			layout.NewSpacer(),
		)
	}

	questionContainer.Content = container.NewVBox(questionWidgets...)
	questionContainer.SetMinSize(fyne.NewSize(600, 500))

	// Lưu câu trả lời vào bảng exam_student_answers mỗi 30 giây
	saveChan := make(chan struct{})
	go func() {
		for {
			select {
			case <-saveChan:
				return
			default:
				time.Sleep(30 * time.Second)
				err := saveAnswersToDB(examStudentID, questions, answers)
				if err != nil {
					log.Printf("Lỗi lưu câu trả lời vào DB: %v", err)
				}
			}
		}
	}()

	stopChan := make(chan struct{})
	closed := false

	closeChannel := func() {
		if !closed {
			close(stopChan)
			close(saveChan)
			closed = true
		}
	}

	go func() {
		for timeLeft > 0 {
			select {
			case <-stopChan:
				return
			default:
				time.Sleep(1 * time.Second)
				timeLeft--
				timer.SetText(formatTime(timeLeft))
				if timeLeft%10 == 0 {
					err := controllers.UpdateRemainingTime(examID, student.ID, timeLeft)
					if err != nil {
						log.Printf("Lỗi cập nhật thời gian còn lại: %v", err)
					}
				}
			}
		}
		closeChannel()
		log.Println("Hết thời gian, tự động nộp bài...")
		submitExam(myApp, myWindow, examID, student, examStudentID, questions, answers)
	}()

	submitButton := widget.NewButton("Nộp bài", func() {
		dialog.ShowConfirm("Xác nhận nộp bài", "Bạn có chắc muốn nộp bài không?", func(confirmed bool) {
			if confirmed {
				closeChannel()
				log.Println("Người dùng xác nhận nộp bài...")
				submitExam(myApp, myWindow, examID, student, examStudentID, questions, answers)
			}
		}, myWindow)
	})

	rightPanel := container.NewVBox(
		answeredLabel,
		questionGrid,
		submitButton,
	)
	rightContainer := container.NewScroll(rightPanel)
	rightContainer.SetMinSize(fyne.NewSize(200, 500))

	topBar := container.NewHBox(
		mssvLabel,
		layout.NewSpacer(),
		timer,
	)

	split := container.NewHSplit(questionContainer, rightContainer)
	split.SetOffset(0.75)

	mainContainer := container.NewBorder(
		container.NewVBox(topBar, canvas.NewLine(theme.Color(theme.ColorNameForeground))),
		nil, nil, nil, split,
	)

	myWindow.SetContent(mainContainer)
	myWindow.SetOnClosed(func() {
		closeChannel()
		// Cập nhật thời gian còn lại khi đóng ứng dụng
		err := controllers.UpdateRemainingTime(examID, student.ID, timeLeft)
		if err != nil {
			log.Printf("Lỗi cập nhật thời gian còn lại khi đóng ứng dụng: %v", err)
		}
		// Lưu câu trả lời cuối cùng khi đóng ứng dụng
		err = saveAnswersToDB(examStudentID, questions, answers)
		if err != nil {
			log.Printf("Lỗi lưu câu trả lời khi đóng ứng dụng: %v", err)
		}
	})
	myWindow.Show()
}

func saveAnswersToDB(examStudentID int, questions []models.Question, answers map[int]string) error {
	for questionID, selected := range answers {
		// Tìm câu trả lời đúng để tính is_correct
		var correctAnswer string
		for _, question := range questions {
			if question.ID == questionID {
				correctAnswer = question.CorrectAnswer
				break
			}
		}
		isCorrect := 0
		if selected == correctAnswer {
			isCorrect = 1
		}

		// Sử dụng INSERT ... ON DUPLICATE KEY UPDATE để upsert
		_, err := database.DB.Exec(`
			INSERT INTO exam_student_answers (exam_student_id, question_id, selected_answer, is_correct, answered_at)
			VALUES (?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
			selected_answer = VALUES(selected_answer),
			is_correct = VALUES(is_correct),
			answered_at = VALUES(answered_at)`,
			examStudentID, questionID, selected, isCorrect, time.Now())
		if err != nil {
			return fmt.Errorf("lỗi lưu câu trả lời vào exam_student_answers: %v", err)
		}
	}

	// Xóa các câu trả lời đã bị bỏ chọn
	for _, question := range questions {
		if _, exists := answers[question.ID]; !exists {
			_, err := database.DB.Exec(`
				DELETE FROM exam_student_answers 
				WHERE exam_student_id = ? AND question_id = ?`,
				examStudentID, question.ID)
			if err != nil {
				return fmt.Errorf("lỗi xóa câu trả lời từ exam_student_answers: %v", err)
			}
		}
	}

	return nil
}

func submitExam(myApp fyne.App, myWindow fyne.Window, examID int, student models.Student, examStudentID int, questions []models.Question, answers map[int]string) {
	log.Println("Bắt đầu nộp bài...")

	// Lưu câu trả lời cuối cùng trước khi nộp bài
	err := saveAnswersToDB(examStudentID, questions, answers)
	if err != nil {
		log.Printf("Lỗi lưu câu trả lời cuối cùng: %v", err)
	}

	var answersList []models.Answer
	for questionID, selected := range answers {
		var correctAnswer string
		for _, question := range questions {
			if question.ID == questionID {
				correctAnswer = question.CorrectAnswer
				break
			}
		}
		answersList = append(answersList, models.Answer{
			QuestionID:    questionID,
			Selected:      selected,
			CorrectAnswer: correctAnswer,
		})
	}

	correctCount := 0
	for _, question := range questions {
		if selected, exists := answers[question.ID]; exists && selected == question.CorrectAnswer {
			correctCount++
		}
	}

	// Gọi SubmitExam để tính điểm và lưu kết quả
	_, err = controllers.SubmitExam(examID, student.ID, questions, answers)
	if err != nil {
		log.Printf("Lỗi nộp bài: %v", err)
		errorWindow := myApp.NewWindow("Lỗi")
		errorWindow.Resize(fyne.NewSize(300, 200))
		errorWindow.SetContent(widget.NewLabel(fmt.Sprintf("Lỗi nộp bài: %v", err)))
		errorWindow.Show()
		myWindow.Close()
		return
	}

	// Cập nhật trạng thái trong exam_students thành 'completed'
	err = controllers.CompleteExam(examID, student.ID)
	if err != nil {
		log.Printf("Lỗi cập nhật trạng thái exam_students: %v", err)
		errorWindow := myApp.NewWindow("Lỗi")
		errorWindow.Resize(fyne.NewSize(300, 200))
		errorWindow.SetContent(widget.NewLabel(fmt.Sprintf("Lỗi cập nhật trạng thái: %v", err)))
		errorWindow.Show()
		myWindow.Close()
		return
	}

	log.Println("Hiển thị trang kết quả...")
	ShowResult(myApp, student, answersList, correctCount, len(questions))

	log.Println("Đóng cửa sổ làm bài...")
	myWindow.Close()
}

func formatTime(seconds int) string {
	minutes := seconds / 60
	secs := seconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, secs)
}
