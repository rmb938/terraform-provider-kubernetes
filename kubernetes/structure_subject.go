package kubernetes

import (
	api "k8s.io/api/rbac/v1"
)

func expandSubjects(subjectsData []interface{}) []api.Subject {
	var subjects []api.Subject

	for _, s := range subjectsData {
		subjectData := s.(map[string]interface{})
		subject := api.Subject{
			APIGroup:  subjectData["api_group"].(string),
			Kind:      subjectData["kind"].(string),
			Name:      subjectData["name"].(string),
			Namespace: subjectData["namespace"].(string),
		}
		subjects = append(subjects, subject)
	}
	return subjects
}

func flattenSubjects(subjects []api.Subject) []map[string]interface{} {
	var subjectsData []map[string]interface{}
	for _, subject := range subjects {
		subjectsData = append(subjectsData, map[string]interface{}{
			"api_group": subject.APIGroup,
			"kind":      subject.Kind,
			"name":      subject.Name,
			"namespace": subject.Namespace,
		})
	}
	return subjectsData
}
