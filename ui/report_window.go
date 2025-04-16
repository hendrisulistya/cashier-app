package ui

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	datepicker "github.com/sdassow/fyne-datepicker"
)

type ReportWindow struct {
	window   fyne.Window
	database *sql.DB
}

func NewReportWindow(window fyne.Window, database *sql.DB) *ReportWindow {
	return &ReportWindow{
		window:   window,
		database: database,
	}
}

func (r *ReportWindow) Load() error {
	content := r.createReportContent()
	r.window.SetContent(content)
	return nil
}

func (r *ReportWindow) createReportContent() fyne.CanvasObject {
	// Back button with icon
	backButton := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		mainWindow := NewMainWindow(r.window, r.database)
		if err := mainWindow.Load(); err != nil {
			log.Printf("Error returning to main menu: %v", err)
		}
	})

	// Header with title
	title := widget.NewLabelWithStyle("Sales Report Generator", fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: true})
	header := container.NewHBox(
		backButton,
		layout.NewSpacer(),
		title,
		layout.NewSpacer(),
	)

	// Date range selection with date picker
	startDate := widget.NewEntry()
	startDate.Resize(fyne.NewSize(400, startDate.MinSize().Height))
	startDate.SetPlaceHolder("Select start date (YYYY-MM-DD)")
	startDateBtn := widget.NewButton("Select Date", func() {
		picker := datepicker.NewDatePicker(time.Now(), time.Sunday, func(date time.Time, selected bool) {
			if selected {
				startDate.SetText(date.Format("2006-01-02"))
			}
		})
		dialog.ShowCustom("Select Start Date", "Done", picker, r.window)
	})

	endDate := widget.NewEntry()
	endDate.Resize(fyne.NewSize(400, endDate.MinSize().Height))
	endDate.SetPlaceHolder("Select end date (YYYY-MM-DD)")
	endDateBtn := widget.NewButton("Select Date", func() {
		picker := datepicker.NewDatePicker(time.Now(), time.Sunday, func(date time.Time, selected bool) {
			if selected {
				endDate.SetText(date.Format("2006-01-02"))
			}
		})
		dialog.ShowCustom("Select End Date", "Done", picker, r.window)
	})

	// Today button with icon
	todayButton := widget.NewButtonWithIcon("Today", theme.MediaPlayIcon(), func() {
		today := time.Now().Format("2006-01-02")
		startDate.SetText(today)
		endDate.SetText(today)
	})
	todayButton.Importance = widget.MediumImportance

	// Date form with better spacing and vertical layout
	dateForm := container.NewVBox(
		container.NewHBox(
			widget.NewLabelWithStyle("Start Date:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			startDate,
			startDateBtn,
		),
		container.NewHBox(
			widget.NewLabelWithStyle("End Date:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			endDate,
			endDateBtn,
		),
		container.NewHBox(
			layout.NewSpacer(),
			todayButton,
			layout.NewSpacer(),
		),
	)

	// Report content with monospace font
	reportText := widget.NewTextGrid()
	reportText.SetStyleRange(0, 0, 0, 0, &widget.CustomTextGridStyle{
		TextStyle: fyne.TextStyle{Monospace: true},
	})

	// Action buttons with icons
	generateButton := widget.NewButtonWithIcon("Generate Report", theme.DocumentCreateIcon(), func() {
		start, err := time.Parse("2006-01-02", startDate.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid start date format"), r.window)
			return
		}

		end, err := time.Parse("2006-01-02", endDate.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid end date format"), r.window)
			return
		}

		report, err := r.generateSalesReport(start, end)
		if err != nil {
			dialog.ShowError(err, r.window)
			return
		}

		reportText.SetText(report)
	})
	generateButton.Importance = widget.HighImportance

	exportButton := widget.NewButtonWithIcon("Export to CSV", theme.DocumentSaveIcon(), func() {
		if startDate.Text == "" || endDate.Text == "" {
			dialog.ShowError(fmt.Errorf("please select both start and end dates"), r.window)
			return
		}

		start, err := time.Parse("2006-01-02", startDate.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid start date format"), r.window)
			return
		}

		end, err := time.Parse("2006-01-02", endDate.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid end date format"), r.window)
			return
		}

		// Query to get sales data
		query := `
			SELECT
				p.name,
				SUM(si.quantity) as total_quantity,
				SUM(si.quantity * p.price) as total_sales
			FROM sales s
			JOIN sale_items si ON s.id = si.sale_id
			JOIN products p ON si.product_id = p.id
			WHERE s.created_at BETWEEN $1 AND $2
			GROUP BY p.name
			ORDER BY total_sales DESC
		`

		rows, err := r.database.Query(query, start, end)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to query sales data: %v", err), r.window)
			return
		}
		defer rows.Close()

		// Create CSV content
		csvContent := fmt.Sprintf("Sales Report from %s to %s\n\n", start.Format("2006-01-02"), end.Format("2006-01-02"))
		csvContent += "Product Name,Quantity,Total Sales\n"

		var totalRevenue float64
		for rows.Next() {
			var (
				name       string
				quantity   int
				totalSales float64
			)
			if err := rows.Scan(&name, &quantity, &totalSales); err != nil {
				dialog.ShowError(fmt.Errorf("failed to scan row: %v", err), r.window)
				return
			}
			csvContent += fmt.Sprintf("%s,%d,Rp%.2f\n", name, quantity, totalSales)
			totalRevenue += totalSales
		}

		csvContent += fmt.Sprintf("\nTotal Revenue,Rp%.2f\n", totalRevenue)

		dialog.ShowInformation("Report Generated", csvContent, r.window)
	})
	exportButton.Importance = widget.MediumImportance

	// Button container with spacing
	buttons := container.NewHBox(
		layout.NewSpacer(),
		generateButton,
		widget.NewLabel(""), // spacing
		exportButton,
		layout.NewSpacer(),
	)

	// Card-like container for the report
	reportCard := container.NewPadded(
		container.NewVBox(
			widget.NewLabelWithStyle("Report Output", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			container.NewPadded(container.NewScroll(reportText)),
		),
	)

	// Main layout with proper spacing and padding
	content := container.NewBorder(
		container.NewVBox(
			header,
			widget.NewSeparator(),
			container.NewPadded(dateForm),
			buttons,
		),
		nil,
		nil,
		nil,
		reportCard,
	)

	return container.NewPadded(content)
}

func (r *ReportWindow) generateSalesReport(start, end time.Time) (string, error) {
	// Query to get sales data
	query := `
        SELECT
            p.name,
            SUM(si.quantity) as total_quantity,
            SUM(si.quantity * si.price_at_sale) as total_sales
        FROM sales s
        JOIN sale_items si ON s.id = si.sale_id
        JOIN products p ON si.product_id = p.id
        WHERE s.created_at BETWEEN $1 AND $2
        GROUP BY p.name
        ORDER BY total_sales DESC
    `

	rows, err := r.database.Query(query, start, end)
	if err != nil {
		return "", fmt.Errorf("failed to query sales data: %v", err)
	}
	defer rows.Close()

	var report string
	report += fmt.Sprintf("Sales Report from %s to %s\n\n", start.Format("2006-01-02"), end.Format("2006-01-02"))
	report += "Product Name | Quantity | Total Sales\n"
	report += "----------------------------------------\n"

	var totalRevenue float64
	for rows.Next() {
		var (
			name       string
			quantity   int
			totalSales float64
		)
		if err := rows.Scan(&name, &quantity, &totalSales); err != nil {
			return "", fmt.Errorf("failed to scan row: %v", err)
		}
		report += fmt.Sprintf("%-20s | %8d | Rp%.2f\n", name, quantity, totalSales)
		totalRevenue += totalSales
	}

	report += "----------------------------------------\n"
	report += fmt.Sprintf("Total Revenue: Rp%.2f\n", totalRevenue)

	return report, nil
}
