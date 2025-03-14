package views

import (
	"fmt"
	"myapp/controllers"
	"myapp/database"
	"myapp/models"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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

		// Gọi hàm LoginStudent và kiểm tra kết quả
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

		// Hiển thị popup thông báo đăng nhập thành công
		showLoginSuccessPopup(myApp, myWindow, *student)
	})

	loginForm := container.NewVBox(
		widget.NewLabel("Đăng nhập hệ thống trắc nghiệm"),
		studentCodeEntry,
		loginButton,
		logLabel,
	)

	myWindow.SetContent(loginForm)
	myWindow.ShowAndRun()
}

func showLoginSuccessPopup(myApp fyne.App, parent fyne.Window, student models.Student) {
	// Lấy thông tin kỳ thi từ exam_students dựa trên student_id
	rows, err := database.DB.Query(`
		SELECT e.name, es.test_started_at, ec.duration, e.active, e.id 
		FROM exam_students es 
		JOIN exams e ON es.exam_id = e.id 
		JOIN exam_config ec ON e.id = ec.exam_id 
		WHERE es.student_id = ? AND es.status = 'pending'`, student.ID)
	if err != nil {
		fmt.Println("Lỗi truy vấn SQL:", err)
		dialog.ShowError(err, parent)
		return
	}
	defer rows.Close()

	var examName string
	var startTime time.Time
	var duration int
	var active bool
	var examID int
	if rows.Next() {
		err := rows.Scan(&examName, &startTime, &duration, &active, &examID)
		if err != nil {
			fmt.Println("Lỗi scan dữ liệu:", err)
			dialog.ShowError(err, parent)
			return
		}
	} else {
		fmt.Println("Không tìm thấy kỳ thi cho student_id:", student.ID)
		dialog.ShowInformation("Thông báo", "Không tìm thấy kỳ thi cho sinh viên này!", parent)
		return
	}

	// Tính thời gian kết thúc (chỉ để hiển thị)
	endTime := startTime.Add(time.Duration(duration) * time.Minute)

	// Chuỗi thông báo
	message := fmt.Sprintf(
		"Đăng nhập thành công!\n\n"+
			"Mã số sinh viên: %s\n"+
			"Tên thí sinh: %s\n"+
			"Lớp: %s\n"+
			"Khoa: %s\n\n"+
			"Kỳ thi: %s\n"+
			"Thời gian làm bài: %s đến %s",
		student.StudentCode, student.Name, student.Class, student.Faculty,
		examName, startTime.Format("15:04 02/01/2006"), endTime.Format("15:04 02/01/2006"),
	)

	// Tạo nội dung cho dialog
	content := container.NewVBox(
		widget.NewLabel(message),
		layout.NewSpacer(),
	)

	// Khai báo popup trước
	var popup *dialog.CustomDialog

	// Thêm nút "Vào làm bài" nếu trạng thái active = 1
	if active {
		fmt.Println("Kỳ thi đang active, hiển thị nút 'Vào làm bài'")
		startButton := widget.NewButton("Vào làm bài", func() {
			popup.Hide()
			ShowExam(myApp, student, examID)
		})
		content.Add(startButton)
	} else {
		fmt.Println("Kỳ thi không active, không hiển thị nút 'Vào làm bài'")
	}

	// Tạo popup
	popup = dialog.NewCustom(
		"Thông tin đăng nhập",
		"Quay lại",
		content,
		parent,
	)

	popup.SetOnClosed(func() {
		// Không làm gì khi quay lại
	})

	popup.Show()
}
