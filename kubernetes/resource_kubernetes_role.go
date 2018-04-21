package kubernetes

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pkgApi "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	api "k8s.io/api/rbac/v1"
)

func resourceKubernetesRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceKubernetesRoleCreate,
		Read:   resourceKubernetesRoleRead,
		Exists: resourceKubernetesRoleExists,
		Update: resourceKubernetesRoleUpdate,
		Delete: resourceKubernetesRoleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"metadata": namespacedMetadataSchema("role", true),
			"rule": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_groups": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"resources": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"resource_names": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"non_resource_urls": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"verbs": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func resourceKubernetesRoleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)
	metadata := expandMetadata(d.Get("metadata").([]interface{}))

	rulesData := d.Get("rule").([]interface{})

	role := &api.Role{
		ObjectMeta: metadata,
		Rules:      expandRules(rulesData),
	}

	out, err := conn.RbacV1().Roles(metadata.Namespace).Create(role)
	if err != nil {
		return err
	}
	d.SetId(buildId(out.ObjectMeta))

	return resourceKubernetesRoleRead(d, meta)
}

func resourceKubernetesRoleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return err
	}

	role, err := conn.RbacV1().Roles(namespace).Get(name, metaV1.GetOptions{})
	err = d.Set("metadata", flattenMetadata(role.ObjectMeta))
	if err != nil {
		return err
	}

	d.Set("rule", flattenRules(role.Rules))

	return nil
}

func resourceKubernetesRoleExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	conn := meta.(*kubernetes.Clientset)

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return false, err
	}
	_, err = conn.RbacV1beta1().Roles(namespace).Get(name, metaV1.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
	}
	return true, err
}

func resourceKubernetesRoleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return err
	}

	ops := patchMetadata("metadata.0.", "/metadata/", d)

	if d.HasChange("rule") {
		rulesData := d.Get("rule").([]interface{})
		ops = append(ops, &ReplaceOperation{
			Path:  "/rules",
			Value: expandRules(rulesData),
		})
	}

	data, err := ops.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Failed to marshal update operations: %s", err)
	}
	out, err := conn.RbacV1beta1().RoleBindings(namespace).Patch(name, pkgApi.JSONPatchType, data)
	if err != nil {
		return fmt.Errorf("Failed to update role: %s", err)
	}
	d.SetId(buildId(out.ObjectMeta))

	return resourceKubernetesRoleRead(d, meta)
}

func resourceKubernetesRoleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return err
	}

	err = conn.RbacV1beta1().Roles(namespace).Delete(name, &metaV1.DeleteOptions{})
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}
