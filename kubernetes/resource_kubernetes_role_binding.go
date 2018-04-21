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

func resourceKubernetesRoleBinding() *schema.Resource {
	return &schema.Resource{
		Create: resourceKubernetesRoleBindingCreate,
		Read:   resourceKubernetesRoleBindingRead,
		Exists: resourceKubernetesRoleBindingExists,
		Update: resourceKubernetesRoleBindingUpdate,
		Delete: resourceKubernetesRoleBindingDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"metadata": namespacedMetadataSchema("role binding", true),
			"role_ref": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_group": {
							Type:     schema.TypeString,
							Required: true,
						},
						"kind": {
							Type:     schema.TypeString,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"subject": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_group": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"kind": {
							Type:     schema.TypeString,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"namespace": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceKubernetesRoleBindingCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)
	metadata := expandMetadata(d.Get("metadata").([]interface{}))

	roleRef := d.Get("role_ref").([]interface{})[0].(map[string]interface{})
	subjects := d.Get("subject").([]interface{})

	roleBinding := &api.RoleBinding{
		ObjectMeta: metadata,
		RoleRef:    expandRoleRef(roleRef),
		Subjects:   expandSubjects(subjects),
	}

	out, err := conn.RbacV1().RoleBindings(metadata.Namespace).Create(roleBinding)
	if err != nil {
		return err
	}
	d.SetId(buildId(out.ObjectMeta))

	return resourceKubernetesRoleBindingRead(d, meta)
}

func resourceKubernetesRoleBindingRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return err
	}

	roleBinding, err := conn.RbacV1().RoleBindings(namespace).Get(name, metaV1.GetOptions{})
	err = d.Set("metadata", flattenMetadata(roleBinding.ObjectMeta))
	if err != nil {
		return err
	}

	d.Set("role_ref", []map[string]interface{}{flattenRoleRef(&roleBinding.RoleRef)})
	d.Set("subject", flattenSubjects(roleBinding.Subjects))

	return nil
}

func resourceKubernetesRoleBindingExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	conn := meta.(*kubernetes.Clientset)

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return false, err
	}
	_, err = conn.RbacV1beta1().RoleBindings(namespace).Get(name, metaV1.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
	}
	return true, err
}

func resourceKubernetesRoleBindingUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return err
	}

	ops := patchMetadata("metadata.0.", "/metadata/", d)

	if d.HasChange("role_ref") {
		roleRef := d.Get("role_ref").([]interface{})[0].(map[string]interface{})
		ops = append(ops, &ReplaceOperation{
			Path:  "/roleRef",
			Value: expandRoleRef(roleRef),
		})
	}
	if d.HasChange("subject") {
		subjects := d.Get("subject").([]interface{})
		ops = append(ops, &ReplaceOperation{
			Path:  "/subjects",
			Value: expandSubjects(subjects),
		})
	}

	data, err := ops.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Failed to marshal update operations: %s", err)
	}
	out, err := conn.RbacV1beta1().RoleBindings(namespace).Patch(name, pkgApi.JSONPatchType, data)
	if err != nil {
		return fmt.Errorf("Failed to update role binding: %s", err)
	}
	d.SetId(buildId(out.ObjectMeta))

	return resourceKubernetesRoleBindingRead(d, meta)
}

func resourceKubernetesRoleBindingDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return err
	}

	err = conn.RbacV1beta1().RoleBindings(namespace).Delete(name, &metaV1.DeleteOptions{})
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}
