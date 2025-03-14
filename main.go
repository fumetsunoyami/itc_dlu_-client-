package main

import (
	"myapp/database"
	"myapp/views"

	"fyne.io/fyne/v2/app"
)

func main() {
	database.ConnectDB() // Kết nối CSDL
	myApp := app.New()
	views.ShowLogin(myApp)
}
