 
package filtering

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"syslog-analyzer/models"
)

// Aggregator handles log aggregation operations
type Aggregator struct {
	rules       []models.AggregationRule
	groups      map[string]*AggregationGroup
	groupsMutex sync.RWMutex
}

// AggregationGroup represents a group of aggregated events
type AggregationGroup struct {
	Key       string
	Events    []models.LogEvent
	FirstSeen time.Time
	LastSeen  time.Time
	Count     int
}

// NewAggregator creates a new aggregation engine
func NewAggregator(rules []models.AggregationRule) *Aggregator {
	return &Aggregator{
		rules:  rules,
		groups: make(map[string]*AggregationGroup),
	}
}

// ProcessBatch processes a batch of events through the aggregation engine
func (a *Aggregator) ProcessBatch(events []models.LogEvent) []models.LogEvent {
	if len(a.rules) == 0 {
		return events // No aggregation needed
	}
	
	// For simplicity, we'll use the first aggregation rule
	// In production, you might want to support multiple rules
	if len(a.rules) > 0 {
		return a.aggregateByRule(events, a.rules[0])
	}
	
	return events
}

// aggregateByRule aggregates events according to a specific rule
func (a *Aggregator) aggregateByRule(events []models.LogEvent, rule models.AggregationRule) []models.LogEvent {
	a.groupsMutex.Lock()
	defer a.groupsMutex.Unlock()
	
	// Clean up old groups based on time window
	a.cleanupOldGroups(rule.TimeWindow)
	
	// Group events
	for _, event := range events {
		groupKey := a.generateGroupKey(event, rule.GroupBy)
		
		if group, exists := a.groups[groupKey]; exists {
			// Add to existing group
			group.Events = append(group.Events, event)
			group.LastSeen = event.Time
			group.Count++
		} else {
			// Create new group
			a.groups[groupKey] = &AggregationGroup{
				Key:       groupKey,
				Events:    []models.LogEvent{event},
				FirstSeen: event.Time,
				LastSeen:  event.Time,
				Count:     1,
			}
		}
	}
	
	// Return aggregated events (one per group)
	var aggregated []models.LogEvent
	for _, group := range a.groups {
		if group.Count > 0 {
			// Create aggregated event
			aggregatedEvent := models.LogEvent{
				Time:   group.LastSeen,
				Source: group.Events[0].Source,
				Event: map[string]interface{}{
					"aggregated_message": fmt.Sprintf("Aggregated %d similar events", group.Count),
					"first_seen":         group.FirstSeen,
					"last_seen":          group.LastSeen,
					"count":              group.Count,
					"group_key":          group.Key,
					"sample_event":       group.Events[0].Event,
				},
				Size: group.Events[0].Size,
			}
			aggregated = append(aggregated, aggregatedEvent)
		}
	}
	
	return aggregated
}

// generateGroupKey generates a unique key for grouping events
func (a *Aggregator) generateGroupKey(event models.LogEvent, groupByFields []string) string {
	groupData := make(map[string]interface{})
	
	for _, field := range groupByFields {
		value := a.extractFieldValue(event, field)
		groupData[field] = value
	}
	
	// Create hash of group data
	dataBytes, _ := json.Marshal(groupData)
	hash := md5.Sum(dataBytes)
	return fmt.Sprintf("%x", hash)
}

// extractFieldValue extracts a field value from a log event (similar to filtering engine)
func (a *Aggregator) extractFieldValue(event models.LogEvent, field string) interface{} {
	switch field {
	case "message":
		return event.Event
	case "source":
		return event.Source
	case "time":
		return event.Time
	default:
		// Try to extract from JSON object
		if jsonObj, ok := event.Event.(map[string]interface{}); ok {
			if value, exists := jsonObj[field]; exists {
				return value
			}
		}
	}
	
	return nil
}

// cleanupOldGroups removes groups that are older than the time window
func (a *Aggregator) cleanupOldGroups(timeWindow time.Duration) {
	cutoff := time.Now().Add(-timeWindow)
	
	for key, group := range a.groups {
		if group.LastSeen.Before(cutoff) {
			delete(a.groups, key)
		}
	}
}