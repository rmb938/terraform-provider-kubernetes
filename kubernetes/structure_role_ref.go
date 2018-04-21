package kubernetes

import (
	api "k8s.io/api/rbac/v1"
)

func expandRoleRef(roleRefData map[string]interface{}) api.RoleRef {
	return api.RoleRef{
		APIGroup: roleRefData["api_group"].(string),
		Kind:     roleRefData["kind"].(string),
		Name:     roleRefData["name"].(string),
	}
}

func flattenRoleRef(roleRef *api.RoleRef) map[string]interface{} {
	return map[string]interface{}{
		"api_group": roleRef.APIGroup,
		"kind":      roleRef.Kind,
		"name":      roleRef.Name,
	}
}
