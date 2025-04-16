package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type LoginPage struct {
	window   fyne.Window
	username *widget.Entry
	password *widget.Entry
	onLogin  func(username, password string) bool
}

func NewLoginPage(window fyne.Window, loginCallback func(username, password string) bool) *LoginPage {
	return &LoginPage{
		window:  window,
		onLogin: loginCallback,
	}
}

func (l *LoginPage) attemptLogin() {
	if l.onLogin(l.username.Text, l.password.Text) {
		l.username.SetText("")
		l.password.SetText("")
	} else {
		// Error message with dark text for visibility
		errorText := canvas.NewText("Invalid username or password", color.NRGBA{R: 200, G: 30, B: 30, A: 255})
		errorText.TextStyle.Bold = true

		content := container.NewCenter(
			container.NewVBox(
				canvas.NewImageFromFile("assets/logo.svg"),
				container.NewCenter(errorText),
				l.createLoginForm(),
			),
		)
		l.window.SetContent(content)
	}
}

func (l *LoginPage) Load() fyne.CanvasObject {
	// Logo
	logo := canvas.NewImageFromFile("assets/logo.svg")
	logo.SetMinSize(fyne.NewSize(300, 90)) // Increased from 200x60 to 300x90
	logo.FillMode = canvas.ImageFillOriginal
	logo.Resize(fyne.NewSize(300, 90)) // Increased from 200x60 to 300x90

	// Create a card-like container for the login form
	formCard := canvas.NewRectangle(theme.BackgroundColor())
	formCard.SetMinSize(fyne.NewSize(400, 500)) // Increased width to accommodate larger logo

	// Welcome text
	welcomeText := canvas.NewText("Welcome back!", theme.ForegroundColor())
	welcomeText.TextSize = 20
	welcomeText.TextStyle.Bold = true

	// Login Form
	l.username = widget.NewEntry()
	l.username.SetPlaceHolder("Username")
	l.username.Resize(fyne.NewSize(200, 40))

	l.password = widget.NewPasswordEntry()
	l.password.SetPlaceHolder("Password")
	l.password.Resize(fyne.NewSize(200, 40))

	// Handle Enter key for both fields
	l.username.OnSubmitted = func(string) {
		l.password.FocusGained()
	}

	l.password.OnSubmitted = func(string) {
		l.attemptLogin()
	}

	// Style the login button
	loginBtn := widget.NewButton("Login", l.attemptLogin)
	loginBtn.Importance = widget.HighImportance
	loginBtn.Resize(fyne.NewSize(200, 40))

	// Create a styled container for the form with more spacing
	formContainer := container.NewVBox(
		widget.NewLabel(""), // Extra spacing
		widget.NewLabel(""), // Extra spacing
		container.NewCenter(logo),
		widget.NewLabel(""), // Extra spacing
		container.NewCenter(welcomeText),
		widget.NewLabel(""), // Spacing
		l.username,
		widget.NewLabel(""), // Spacing
		l.password,
		widget.NewLabel(""), // Spacing
		loginBtn,
	)

	// Create a card effect with padding and shadow
	cardContainer := container.NewMax(
		formCard,
		container.NewPadded(formContainer),
	)

	// Set up keyboard shortcuts
	canvas := l.window.Canvas()
	canvas.SetOnTypedKey(func(key *fyne.KeyEvent) {
		if key.Name == fyne.KeyReturn || key.Name == fyne.KeyEnter {
			l.attemptLogin()
		}
	})

	// Main container with centered card
	return container.NewCenter(
		container.NewPadded(cardContainer),
	)
}

func (l *LoginPage) createLoginForm() fyne.CanvasObject {
	loginBtn := widget.NewButton("Login", l.attemptLogin)
	loginBtn.Importance = widget.HighImportance

	return container.NewVBox(
		l.username,
		widget.NewLabel(""), // Spacing
		l.password,
		widget.NewLabel(""), // Spacing
		loginBtn,
	)
}
