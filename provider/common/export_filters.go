package common

import "strings"

// ConvertFiltersToMeld converts the Terraform filter representation to the Funnel Meld format.
func ConvertFiltersToMeld(filters []ExportFilterJSON) map[string]any {
	if len(filters) == 0 {
		return nil
	}

	filterConditions := make([]map[string]any, 0, len(filters))
	for _, filter := range filters {
		condition := map[string]any{
			filter.FieldId: buildToMeldFieldCondition(filter),
		}
		filterConditions = append(filterConditions, condition)
	}

	return map[string]any{
		"=and": filterConditions,
	}
}

// When converting to Meld, handle that fields can contain an array of OR statements or a single operation and value.
func buildToMeldFieldCondition(filter ExportFilterJSON) map[string]any {
	if len(filter.Or) > 0 {
		orConditions := make([]map[string]any, 0, len(filter.Or))
		for _, orItem := range filter.Or {
			orConditions = append(orConditions, map[string]any{
				"=" + orItem.Operation: orItem.Value,
			})
		}

		return map[string]any{
			"=or": orConditions,
		}
	}

	return map[string]any{
		"=" + filter.Operation: filter.Value,
	}
}

// ConvertFiltersFromMeld converts the Funnel Meld format to the Terraform filter representation.
func ConvertFiltersFromMeld(filters map[string]any) []ExportFilterJSON {
	var andList []any

	if val, ok := filters["=and"]; ok {
		andList = val.([]any)
	} else {
		andList = []any{filters}
	}

	result := make([]ExportFilterJSON, 0, len(andList))
	for _, condition := range andList {
		conditionMap, ok := condition.(map[string]any)
		if !ok {
			continue
		}

		for fieldId, fieldCondition := range conditionMap {
			fieldConditionMap, ok := fieldCondition.(map[string]any)
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
func buildFromMeldFieldCondition(fieldId string, conditions map[string]any) ExportFilterJSON {
	filter := ExportFilterJSON{FieldId: fieldId}

	if orList, hasOr := conditions["=or"].([]any); hasOr {
		for _, item := range orList {
			if itemMap, ok := item.(map[string]any); ok {
				for key, val := range itemMap {
					op := strings.TrimPrefix(key, "=")
					filter.Or = append(filter.Or, ExportFilterOrJSON{Operation: op, Value: val.(string)})
				}
			}
		}
	} else {
		for key, val := range conditions {
			if after, ok := strings.CutPrefix(key, "="); ok {
				filter.Operation = after
				filter.Value = val.(string)
				break
			}
		}
	}

	return filter
}
