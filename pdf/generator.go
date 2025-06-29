package pdf

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/jung-kurt/gofpdf"

	"syslog-analyzer/models"
)

// Generator handles PDF report generation
type Generator struct {
	pdf *gofpdf.Fpdf
}

// NewGenerator creates a new PDF generator
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateReport generates a comprehensive PDF report
func (g *Generator) GenerateReport(sources []models.SourceMetrics, global models.GlobalMetrics) ([]byte, error) {
	// Initialize PDF
	g.pdf = gofpdf.New("P", "mm", "A4", "")
	g.pdf.SetMargins(20, 20, 20)
	g.pdf.SetAutoPageBreak(true, 25)
	
	// Add page
	g.pdf.AddPage()
	
	// Generate report content
	g.addHeader()
	g.addGlobalSummary(global)
	g.addSourcesOverview(sources)
	g.addDetailedSourceMetrics(sources)
	g.addFooter()
	
	// Check for errors
	if err := g.pdf.Error(); err != nil {
		return nil, fmt.Errorf("PDF generation error: %v", err)
	}
	
	// Get PDF bytes using buffer
	var buf bytes.Buffer
	err := g.pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to output PDF: %v", err)
	}
	
	return buf.Bytes(), nil
}

// addHeader adds the report header
func (g *Generator) addHeader() {
	// Title
	g.pdf.SetFont("Arial", "B", 24)
	g.pdf.SetTextColor(44, 62, 80) // Dark blue
	g.pdf.CellFormat(0, 15, "Syslog Analyzer Report", "", 1, "C", false, 0, "")
	
	// Subtitle with generation time
	g.pdf.SetFont("Arial", "", 12)
	g.pdf.SetTextColor(127, 140, 141) // Gray
	generatedAt := time.Now().Format("January 2, 2006 at 15:04:05 MST")
	g.pdf.CellFormat(0, 10, fmt.Sprintf("Generated on %s", generatedAt), "", 1, "C", false, 0, "")
	
	// Add some space
	g.pdf.Ln(10)
}

// addGlobalSummary adds the global metrics summary
func (g *Generator) addGlobalSummary(global models.GlobalMetrics) {
	// Section header
	g.pdf.SetFont("Arial", "B", 16)
	g.pdf.SetTextColor(52, 73, 94) // Dark blue
	g.pdf.CellFormat(0, 10, "Global Summary", "", 1, "L", false, 0, "")
	g.pdf.Ln(5)
	
	// Create summary table
	g.pdf.SetFont("Arial", "", 11)
	g.pdf.SetTextColor(0, 0, 0)
	
	// Table headers
	g.pdf.SetFillColor(231, 243, 250) // Light blue
	g.pdf.SetFont("Arial", "B", 10)
	g.pdf.CellFormat(60, 8, "Metric", "1", 0, "L", true, 0, "")
	g.pdf.CellFormat(40, 8, "Value", "1", 1, "R", true, 0, "")
	
	// Table data
	g.pdf.SetFont("Arial", "", 10)
	g.pdf.SetFillColor(255, 255, 255) // White
	
	metrics := []struct {
		name  string
		value string
	}{
		{"Total Real-time EPS", fmt.Sprintf("%.2f", global.TotalRealTimeEPS)},
		{"Total Real-time GB/s", fmt.Sprintf("%.6f", global.TotalRealTimeGBps)},
		{"Total Logs Ingested", formatNumber(global.TotalLogsIngested)},
		{"Hourly Average Logs", formatNumber(global.TotalHourlyAvgLogs)},
		{"Daily Average Logs", formatNumber(global.TotalDailyAvgLogs)},
		{"Active Sources", fmt.Sprintf("%d", global.ActiveSources)},
		{"Total Sources", fmt.Sprintf("%d", global.TotalSources)},
		{"Total Queue Depth", formatNumber(global.TotalQueueDepth)},
		{"Total Processed", formatNumber(global.TotalProcessedCount)},
		{"Total Sent", formatNumber(global.TotalSentCount)},
	}
	
	for i, metric := range metrics {
		fillColor := i%2 == 0
		g.pdf.CellFormat(60, 7, metric.name, "1", 0, "L", fillColor, 0, "")
		g.pdf.CellFormat(40, 7, metric.value, "1", 1, "R", fillColor, 0, "")
	}
	
	g.pdf.Ln(10)
}

// addSourcesOverview adds a high-level overview of all sources
func (g *Generator) addSourcesOverview(sources []models.SourceMetrics) {
	// Section header
	g.pdf.SetFont("Arial", "B", 16)
	g.pdf.SetTextColor(52, 73, 94)
	g.pdf.CellFormat(0, 10, "Sources Overview", "", 1, "L", false, 0, "")
	g.pdf.Ln(5)
	
	if len(sources) == 0 {
		g.pdf.SetFont("Arial", "I", 11)
		g.pdf.SetTextColor(127, 140, 141)
		g.pdf.CellFormat(0, 10, "No sources configured", "", 1, "L", false, 0, "")
		g.pdf.Ln(5)
		return
	}
	
	// Sort sources by name
	sortedSources := make([]models.SourceMetrics, len(sources))
	copy(sortedSources, sources)
	sort.Slice(sortedSources, func(i, j int) bool {
		return sortedSources[i].Name < sortedSources[j].Name
	})
	
	// Check if we need a new page
	if g.pdf.GetY() > 200 {
		g.pdf.AddPage()
	}
	
	// Create sources table
	g.pdf.SetFont("Arial", "B", 9)
	g.pdf.SetFillColor(231, 243, 250)
	
	// Table headers
	headers := []string{"Source Name", "IP:Port", "EPS", "Total Logs", "Status"}
	widths := []float64{40, 35, 25, 30, 30}
	
	for i, header := range headers {
		g.pdf.CellFormat(widths[i], 8, header, "1", 0, "C", true, 0, "")
	}
	g.pdf.Ln(-1)
	
	// Table data
	g.pdf.SetFont("Arial", "", 8)
	
	for i, source := range sortedSources {
		// Check if we need a new page
		if g.pdf.GetY() > 270 {
			g.pdf.AddPage()
		}
		
		fillColor := i%2 == 0
		g.pdf.SetFillColor(255, 255, 255)
		if fillColor {
			g.pdf.SetFillColor(249, 249, 249)
		}
		
		// Determine status
		status := "Inactive"
		if source.IsActive && source.IsReceiving {
			status = "Active"
		} else if source.IsActive {
			status = "Idle"
		}
		
		// Source data
		data := []string{
			truncateString(source.Name, 20),
			fmt.Sprintf("%s:%d", source.SourceIP, source.Port),
			fmt.Sprintf("%.1f", source.RealTimeEPS),
			formatNumber(source.TotalLogsIngested),
			status,
		}
		
		for j, cell := range data {
			g.pdf.CellFormat(widths[j], 7, cell, "1", 0, "C", fillColor, 0, "")
		}
		g.pdf.Ln(-1)
	}
	
	g.pdf.Ln(10)
}

// addDetailedSourceMetrics adds detailed metrics for each source
func (g *Generator) addDetailedSourceMetrics(sources []models.SourceMetrics) {
	if len(sources) == 0 {
		return
	}
	
	// Sort sources by name
	sortedSources := make([]models.SourceMetrics, len(sources))
	copy(sortedSources, sources)
	sort.Slice(sortedSources, func(i, j int) bool {
		return sortedSources[i].Name < sortedSources[j].Name
	})
	
	// Add new page for detailed metrics
	g.pdf.AddPage()
	
	// Section header
	g.pdf.SetFont("Arial", "B", 16)
	g.pdf.SetTextColor(52, 73, 94)
	g.pdf.CellFormat(0, 10, "Detailed Source Metrics", "", 1, "L", false, 0, "")
	g.pdf.Ln(5)
	
	for _, source := range sortedSources {
		// Check if we need a new page
		if g.pdf.GetY() > 240 {
			g.pdf.AddPage()
		}
		
		g.addSourceDetails(source)
	}
}

// addSourceDetails adds detailed metrics for a single source
func (g *Generator) addSourceDetails(source models.SourceMetrics) {
	// Source name header
	g.pdf.SetFont("Arial", "B", 12)
	g.pdf.SetTextColor(44, 62, 80)
	g.pdf.CellFormat(0, 8, fmt.Sprintf("Source: %s", source.Name), "", 1, "L", false, 0, "")
	
	// Source info
	g.pdf.SetFont("Arial", "", 10)
	g.pdf.SetTextColor(127, 140, 141)
	
	protocol := source.Protocol
	if protocol == "" {
		protocol = "N/A"
	}
	
	simulationMode := "OFF"
	if source.SimulationMode {
		simulationMode = "ON"
	}
	
	g.pdf.CellFormat(0, 6, fmt.Sprintf("Address: %s:%d (%s) | Simulation Mode: %s", 
		source.SourceIP, source.Port, protocol, simulationMode), "", 1, "L", false, 0, "")
	g.pdf.Ln(3)
	
	// Metrics table
	g.pdf.SetFont("Arial", "B", 9)
	g.pdf.SetFillColor(231, 243, 250)
	g.pdf.SetTextColor(0, 0, 0)
	
	// Headers
	g.pdf.CellFormat(50, 7, "Metric", "1", 0, "L", true, 0, "")
	g.pdf.CellFormat(30, 7, "Value", "1", 1, "R", true, 0, "")
	
	// Data
	g.pdf.SetFont("Arial", "", 8)
	g.pdf.SetFillColor(255, 255, 255)
	
	metrics := []struct {
		name  string
		value string
	}{
		{"Real-time EPS", fmt.Sprintf("%.2f", source.RealTimeEPS)},
		{"Real-time GB/s", fmt.Sprintf("%.6f", source.RealTimeGBps)},
		{"Total Logs Ingested", formatNumber(source.TotalLogsIngested)},
		{"Hourly Avg Logs", formatNumber(source.HourlyAvgLogs)},
		{"Hourly Avg GB", fmt.Sprintf("%.4f", source.HourlyAvgGB)},
		{"Daily Avg Logs", formatNumber(source.DailyAvgLogs)},
		{"Daily Avg GB", fmt.Sprintf("%.4f", source.DailyAvgGB)},
		{"Queue Depth", formatNumber(source.QueueDepth)},
		{"Processed Count", formatNumber(source.ProcessedCount)},
		{"Sent Count", formatNumber(source.SentCount)},
	}
	
	for i, metric := range metrics {
		fillColor := i%2 == 0
		g.pdf.CellFormat(50, 6, metric.name, "1", 0, "L", fillColor, 0, "")
		g.pdf.CellFormat(30, 6, metric.value, "1", 1, "R", fillColor, 0, "")
	}
	
	// Status and timing info
	g.pdf.Ln(3)
	g.pdf.SetFont("Arial", "", 8)
	g.pdf.SetTextColor(127, 140, 141)
	
	status := "Inactive"
	if source.IsActive && source.IsReceiving {
		status = "Active & Receiving"
	} else if source.IsActive {
		status = "Idle: Waiting for Logs"
	}
	
	lastMessageTime := "Never"
	if !source.LastMessageAt.IsZero() {
		lastMessageTime = source.LastMessageAt.Format("2006-01-02 15:04:05")
	}
	
	g.pdf.CellFormat(0, 5, fmt.Sprintf("Status: %s | Last Message: %s | Last Updated: %s", 
		status, lastMessageTime, source.LastUpdated.Format("2006-01-02 15:04:05")), "", 1, "L", false, 0, "")
	
	g.pdf.Ln(8)
}

// addFooter adds the report footer
func (g *Generator) addFooter() {
	g.pdf.SetY(-15)
	g.pdf.SetFont("Arial", "I", 8)
	g.pdf.SetTextColor(127, 140, 141)
	g.pdf.CellFormat(0, 10, fmt.Sprintf("Generated by Professional Syslog Analyzer - Page %d", g.pdf.PageNo()), "", 0, "C", false, 0, "")
}

// Helper functions

// formatNumber formats large numbers with commas
func formatNumber(n int64) string {
	if n == 0 {
		return "0"
	}
	
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}
	
	// Add commas every 3 digits
	result := ""
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(digit)
	}
	
	return result
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}