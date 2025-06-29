 
package filtering

import (
	"log"
	"regexp"
	"strings"

	"syslog-analyzer/models"
)

// Engine handles log filtering operations
type Engine struct {
	rules []models.FilterRule
}

// NewEngine creates a new filtering engine
func NewEngine(rules []models.FilterRule) *Engine {
	return &Engine{
		rules: rules,
	}
}

// ProcessBatch processes a batch of events through the filter engine
func (e *Engine) ProcessBatch(events []models.LogEvent) []models.LogEvent {
	if len(e.rules) == 0 {
		return events // No filtering needed
	}
	
	var filtered []models.LogEvent
	
	for _, event := range events {
		if e.shouldInclude(event) {
			filtered = append(filtered, event)
		}
	}
	
	return filtered
}

// shouldInclude determines if an event should be included based on filter rules
func (e *Engine) shouldInclude(event models.LogEvent) bool {
	for _, rule := range e.rules {
		match := e.evaluateRule(event, rule)
		
		if rule.Action == "exclude" && match {
			return false // Exclude this event
		}
		if rule.Action == "include" && !match {
			return false // Only include if it matches
		}
	}
	
	return true // Include by default
}

// evaluateRule evaluates a single filter rule against an event
func (e *Engine) evaluateRule(event models.LogEvent, rule models.FilterRule) bool {
	fieldValue := e.extractFieldValue(event, rule.Field)
	if fieldValue == "" {
		return false
	}
	
	switch rule.Operator {
	case "contains":
		return strings.Contains(strings.ToLower(fieldValue), strings.ToLower(rule.Value))
	case "equals":
		return strings.EqualFold(fieldValue, rule.Value)
	case "regex":
		matched, err := regexp.MatchString(rule.Value, fieldValue)
		if err != nil {
			log.Printf("⚠ Invalid regex pattern '%s': %v", rule.Value, err)
			return false
		}
		return matched
	default:
		log.Printf("⚠ Unknown filter operator: %s", rule.Operator)
		return false
	}
}

// extractFieldValue extracts a field value from a log event
func (e *Engine) extractFieldValue(event models.LogEvent, field string) string {
	switch field {
	case "message":
		if str, ok := event.Event.(string); ok {
			return str
		}
		if jsonObj, ok := event.Event.(map[string]interface{}); ok {
			if msg, exists := jsonObj["message"]; exists {
				if msgStr, ok := msg.(string); ok {
					return msgStr
				}
			}
		}
	case "source":
		return event.Source
	case "time":
		return event.Time.String()
	default:
		// Try to extract from JSON object
		if jsonObj, ok := event.Event.(map[string]interface{}); ok {
			if value, exists := jsonObj[field]; exists {
				if strValue, ok := value.(string); ok {
					return strValue
				}
			}
		}
	}
	
	return ""
}