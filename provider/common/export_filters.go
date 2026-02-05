package common

import "strings"

// ConvertFiltersToMeld converts the Terraform filter representation to the Funnel Meld format.
func ConvertFiltersToMeld(filters []ExportFilterJSON) map[string]interface{} {
	if len(filters) == 0 {
		return nil
	}

	filterConditions := make([]map[string]interface{}, 0, len(filters))
	for _, filter := range filters {
		condition := map[string]interface{}{
			filter.FieldId: buildToMeldFieldCondition(filter),
		}
		filterConditions = append(filterConditions, condition)
	}

	return map[string]interface{}{
		"=and": filterConditions,
	}
}

// When converting to Meld, handle that fields can contain an array of OR statements or a single operation and value.
func buildToMeldFieldCondition(filter ExportFilterJSON) map[string]interface{} {
	if len(filter.Or) > 0 {
		orConditions := make([]map[string]interface{}, 0, len(filter.Or))
		for _, orItem := range filter.Or {
			orConditions = append(orConditions, map[string]interface{}{
				"=" + orItem.Operation: orItem.Value,
			})
		}

		return map[string]interface{}{
			"=or": orConditions,
		}
	}

	return map[string]interface{}{
		"=" + filter.Operation: filter.Value,
	}
}

// ConvertFiltersFromMeld converts the Funnel Meld format to the Terraform filter representation.
func ConvertFiltersFromMeld(filters map[string]interface{}) []ExportFilterJSON {
	var andList []interface{}

	if val, ok := filters["=and"]; ok {
		andList = val.([]interface{})
	} else {
		andList = []interface{}{filters}
	}

	result := make([]ExportFilterJSON, 0, len(andList))
	for _, condition := range andList {
		conditionMap, ok := condition.(map[string]interface{})
		if !ok {
			continue
		}

		for fieldId, fieldCondition := range conditionMap {
			fieldConditionMap, ok := fieldCondition.(map[string]interface{})
			if !ok {
				continue
			}

			filter := buildFromMeldFieldCondition(fieldId, fieldConditionMap)
			result = append(result, filter)
		}
	}

	return result
}

// When converting from Meld, handle that fields can contain an array of OR statements or a single operation and value.
func buildFromMeldFieldCondition(fieldId string, conditions map[string]interface{}) ExportFilterJSON {
	filter := ExportFilterJSON{FieldId: fieldId}

	if orList, hasOr := conditions["=or"].([]any); hasOr {
		for _, item := range orList {
			if itemMap, ok := item.(map[string]interface{}); ok {
				for key, val := range itemMap {
					op := strings.TrimPrefix(key, "=")
					filter.Or = append(filter.Or, ExportFilterOrJSON{Operation: op, Value: val.(string)})
				}
			}
		}
	} else {
		for key, val := range conditions {
			if strings.HasPrefix(key, "=") {
				filter.Operation = strings.TrimPrefix(key, "=")
				filter.Value = val.(string)
				break
			}
		}
	}

	return filter
}
