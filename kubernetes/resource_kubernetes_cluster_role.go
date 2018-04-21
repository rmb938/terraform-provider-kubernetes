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

func resourceKubernetesClusterRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceKubernetesClusterRoleCreate,
		Read:   resourceKubernetesClusterRoleRead,
		Exists: resourceKubernetesClusterRoleExists,
		Update: resourceKubernetesClusterRoleUpdate,
		Delete: resourceKubernetesClusterRoleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"metadata": metadataSchema("cluster role", true),
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

func resourceKubernetesClusterRoleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)
	metadata := expandMetadata(d.Get("metadata").([]interface{}))

	rulesData := d.Get("rule").([]interface{})

	role := &api.ClusterRole{
		ObjectMeta: metadata,
		Rules:      expandRules(rulesData),
	}

	out, err := conn.RbacV1().ClusterRoles().Create(role)
	if err != nil {
		return err
	}
	d.SetId(out.Name)

	return resourceKubernetesClusterRoleRead(d, meta)
}

func resourceKubernetesClusterRoleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)

	role, err := conn.RbacV1().ClusterRoles().Get(d.Id(), metaV1.GetOptions{})
	err = d.Set("metadata", flattenMetadata(role.ObjectMeta))
	if err != nil {
		return err
	}

	d.Set("rule", flattenRules(role.Rules))

	return nil
}

func resourceKubernetesClusterRoleExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	conn := meta.(*kubernetes.Clientset)

	_, err := conn.RbacV1beta1().ClusterRoles().Get(d.Id(), metaV1.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
	}
	return true, err
}

func resourceKubernetesClusterRoleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)
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
	out, err := conn.RbacV1beta1().ClusterRoleBindings().Patch(d.Id(), pkgApi.JSONPatchType, data)
	if err != nil {
		return fmt.Errorf("Failed to update cluster role: %s", err)
	}
	d.SetId(out.Name)

	return resourceKubernetesClusterRoleRead(d, meta)
}

func resourceKubernetesClusterRoleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)

	err := conn.RbacV1beta1().ClusterRoles().Delete(d.Id(), &metaV1.DeleteOptions{})
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}
