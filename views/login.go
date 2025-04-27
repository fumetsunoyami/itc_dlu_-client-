package views

import (
	"fmt"
	"log"
	"myapp/controllers"
	"myapp/database"
	"myapp/models"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func ShowLogin(myApp fyne.App) {
	myWindow := myApp.NewWindow("Đăng nhập")
	myWindow.Resize(fyne.NewSize(400, 300))

	studentCodeEntry := widget.NewEntry()
	studentCodeEntry.SetPlaceHolder("Số báo danh/MSSV")

	logLabel := widget.NewLabel("")

	loginButton := widget.NewButton("Đăng nhập", func() {
		studentCode := studentCodeEntry.Text
		if studentCode == "" {
			logLabel.SetText("Vui lòng nhập số báo danh/MSSV")
			return
		}

		// Kiểm tra sinh viên trong bảng students
		var student *models.Student
		var err error
		student, err = controllers.LoginStudent(studentCode)
		if err != nil {
			logLabel.SetText("Lỗi hệ thống, vui lòng thử lại")
			return
		}
		if student == nil {
			logLabel.SetText("Số báo danh/MSSV không đúng!")
			return
		}

		// Nếu sinh viên tồn tại, đóng cửa sổ đăng nhập và hiển thị thông tin
		myWindow.Close()
		ShowStudentInfoWindow(myApp, *student)
	})

	loginForm := container.NewVBox(
		widget.NewLabel("Đăng nhập hệ thống trắc nghiệm"),
		studentCodeEntry,
		loginButton,
		logLabel,
	)

	myWindow.SetContent(loginForm)
	myWindow.Show()
}

func ShowStudentInfoWindow(myApp fyne.App, student models.Student) {
	infoWindow := myApp.NewWindow("Thông tin sinh viên")
	infoWindow.Resize(fyne.NewSize(500, 400))

	// Hiển thị thông tin sinh viên
	message := fmt.Sprintf(
		"Thông tin sinh viên:\n\n"+
			"Số báo danh: %s\n"+
			"Tên: %s\n"+
			"Giới tính: %s\n"+
			"Ngày sinh: %s\n"+
			"Nơi sinh: %s\n\n",
		student.StudentCode, student.Name, student.Gender, student.DateOfBirth, student.PlaceOfBirth,
	)

	// Kiểm tra bảng exam_students để xem sinh viên có kỳ thi nào không
	rows, err := database.DB.Query(`
		SELECT e.name, es.test_started_at, ec.duration, e.id, es.status 
		FROM exam_students es 
		JOIN exams e ON es.exam_id = e.id 
		JOIN exam_config ec ON e.id = ec.exam_id 
		WHERE es.student_id = ?`, student.ID)
	if err != nil {
		fmt.Println("Lỗi truy vấn SQL:", err)
		infoWindow.SetContent(widget.NewLabel("Lỗi tải thông tin kỳ thi"))
		infoWindow.Show()
		return
	}
	defer rows.Close()

	var examName string
	var startTime time.Time
	var duration int
	var examID int
	var status string
	hasExam := false
	if rows.Next() {
		err := rows.Scan(&examName, &startTime, &duration, &examID, &status)
		if err != nil {
			fmt.Println("Lỗi scan dữ liệu:", err)
			infoWindow.SetContent(widget.NewLabel("Lỗi xử lý dữ liệu kỳ thi"))
			infoWindow.Show()
			return
		}
		hasExam = true
	} else {
		fmt.Println("Không tìm thấy kỳ thi cho student_id:", student.ID)
	}

	// Thêm thông tin kỳ thi nếu có
	if hasExam {
		endTime := startTime.Add(time.Duration(duration) * time.Minute)
		examInfo := fmt.Sprintf(
			"Kỳ thi: %s\n"+
				"Thời gian làm bài: %s đến %s\n"+
				"Trạng thái: %s",
			examName,
			startTime.Format("15:04 02/01/2006"),
			endTime.Format("15:04 02/01/2006"),
			status,
		)
		message += examInfo
	} else {
		message += "Không tìm thấy kỳ thi cho sinh viên này!"
	}

	// Tạo nội dung hiển thị
	content := container.NewVBox(
		widget.NewLabel(message),
		layout.NewSpacer(),
	)

	// Kiểm tra trạng thái status để hiển thị nút "Vào làm bài"
	if hasExam && status == "in_progress" {
		fmt.Println("Trạng thái 'in_progress', hiển thị nút 'Vào làm bài'")
		startButton := widget.NewButton("Vào làm bài", func() {
			// Kiểm tra trạng thái kỳ thi
			var status string
			err := database.DB.QueryRow(`
				SELECT status 
				FROM exam_students 
				WHERE exam_id = ? AND student_id = ?`,
				examID, student.ID).Scan(&status)
			if err != nil {
				log.Printf("Lỗi kiểm tra trạng thái: %v", err)
				log.Printf("Lỗi kiểm tra trạng thái kỳ thi")
				return
			}
			if status == "in_progress" {
				infoWindow.Close()
				ShowExam(myApp, student, examID)
			} else {
				log.Printf("Thông báo: Kỳ thi chưa bắt đầu hoặc đã hoàn thành")
			}
		})
		content.Add(startButton)
	} else if hasExam {
		if status == "pending" {
			fmt.Println("Trạng thái 'pending', ẩn nút 'Vào làm bài'")
		} else if status == "completed" {
			fmt.Println("Trạng thái 'completed', ẩn nút 'Vào làm bài'")
		}
	}

	// Nút quay lại
	backButton := widget.NewButton("Quay lại", func() {
		infoWindow.Close()
		ShowLogin(myApp)
	})

	// Sử dụng container.NewHBox thay vì layout.NewHBoxLayout
	hbox := container.NewHBox(layout.NewSpacer(), backButton)

	// Thêm hbox vào đầu content
	content.Objects = append([]fyne.CanvasObject{hbox}, content.Objects...)
	content.Refresh()

	infoWindow.SetContent(content)
	infoWindow.Show()
}
