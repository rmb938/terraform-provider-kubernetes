package kubernetes

import (
	api "k8s.io/api/rbac/v1"
)

func expandRules(rulesData []interface{}) []api.PolicyRule {
	var rules []api.PolicyRule

	for _, s := range rulesData {
		ruleData := s.(map[string]interface{})

		var apiGroups []string
		for _, l := range ruleData["api_groups"].([]interface{}) {
			if l == nil {
				l = ""
			}
			apiGroups = append(apiGroups, l.(string))
		}
		var resources []string
		for _, l := range ruleData["resources"].([]interface{}) {
			if l == nil {
				l = ""
			}
			resources = append(resources, l.(string))
		}
		var resourceNames []string
		for _, l := range ruleData["resource_names"].([]interface{}) {
			if l == nil {
				l = ""
			}
			resourceNames = append(resourceNames, l.(string))
		}
		var nonResourceUrls []string
		for _, l := range ruleData["non_resource_urls"].([]interface{}) {
			if l == nil {
				l = ""
			}
			nonResourceUrls = append(nonResourceUrls, l.(string))
		}
		var verbs []string
		for _, l := range ruleData["verbs"].([]interface{}) {
			if l == nil {
				l = ""
			}
			verbs = append(verbs, l.(string))
		}

		rule := api.PolicyRule{
			APIGroups:       apiGroups,
			Resources:       resources,
			ResourceNames:   resourceNames,
			NonResourceURLs: nonResourceUrls,
			Verbs:           verbs,
		}
		rules = append(rules, rule)
	}

	return rules
}

func flattenRules(rules []api.PolicyRule) []map[string]interface{} {
	var rulesData []map[string]interface{}
	for _, rule := range rules {
		rulesData = append(rulesData, map[string]interface{}{
			"api_groups":        rule.APIGroups,
			"resources":         rule.Resources,
			"resource_names":    rule.ResourceNames,
			"non_resource_urls": rule.NonResourceURLs,
			"verbs":             rule.Verbs,
		})
	}
	return rulesData
}
