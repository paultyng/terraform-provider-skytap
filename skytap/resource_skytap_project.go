package skytap

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/pkg/errors"
	"github.com/skytap/skytap-sdk-go/skytap"
	"github.com/skytap/terraform-provider-skytap/skytap/utils"
)

func resourceSkytapProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceSkytapProjectCreate,
		Read:   resourceSkytapProjectRead,
		Update: resourceSkytapProjectUpdate,
		Delete: resourceSkytapProjectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"summary": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"auto_add_role_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(skytap.ProjectRoleViewer),
					string(skytap.ProjectRoleParticipant),
					string(skytap.ProjectRoleEditor),
					string(skytap.ProjectRoleManager),
				}, false),
			},

			"show_project_members": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceSkytapProjectCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).projectsClient
	ctx := meta.(*SkytapClient).StopContext

	log.Printf("[INFO] preparing arguments for creating the SkyTap Project")

	name := d.Get("name").(string)
	showProjectMembers := d.Get("show_project_members").(bool)

	opts := skytap.Project{
		Name:               &name,
		ShowProjectMembers: &showProjectMembers,
	}

	if v, ok := d.GetOk("summary"); ok {
		opts.Summary = utils.String(v.(string))
	}

	if v, ok := d.GetOk("auto_add_role_name"); ok {
		autoAddRoleName := skytap.ProjectRole(v.(string))
		opts.AutoAddRoleName = &autoAddRoleName
	}

	log.Printf("[DEBUG] Project create options: %#v", opts)
	project, err := client.Create(ctx, &opts)
	if err != nil {
		return errors.Errorf("Error creating project: %v", err)
	}

	d.SetId(*project.Id)

	return resourceSkytapProjectRead(d, meta)
}

func resourceSkytapProjectRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).projectsClient
	ctx := meta.(*SkytapClient).StopContext

	id := d.Id()

	log.Printf("[INFO] Retrieving project: %s", id)
	project, err := client.Get(ctx, id)
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			log.Printf("[DEBUG] Project (%s) was not found - removing from state", id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error retrieving project (%s): %v", id, err)
	}

	d.Set("name", project.Name)
	d.Set("summary", project.Summary)
	d.Set("auto_add_role_name", project.AutoAddRoleName)
	d.Set("show_project_members", project.ShowProjectMembers)

	return err
}

func resourceSkytapProjectUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).projectsClient
	ctx := meta.(*SkytapClient).StopContext

	id := d.Id()
	name := d.Get("name").(string)
	showProjectMembers := d.Get("show_project_members").(bool)

	opts := skytap.Project{
		Name:               &name,
		ShowProjectMembers: &showProjectMembers,
	}

	if v, ok := d.GetOk("summary"); ok {
		opts.Summary = utils.String(v.(string))
	}

	if v, ok := d.GetOk("auto_add_role_name"); ok {
		autoAddRoleName := skytap.ProjectRole(v.(string))
		opts.AutoAddRoleName = &autoAddRoleName
	}

	log.Printf("[DEBUG] Project update options: %#v", opts)
	_, err := client.Update(ctx, id, &opts)
	if err != nil {
		return errors.Errorf("Error updating project (%s): %v", id, err)
	}

	return resourceSkytapProjectRead(d, meta)
}

func resourceSkytapProjectDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).projectsClient
	ctx := meta.(*SkytapClient).StopContext

	id := d.Id()

	log.Printf("[INFO] Destroying project: %s", id)
	err := client.Delete(ctx, id)
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			log.Printf("[DEBUG] Project (%s) was not found - assuming removed", id)
			return nil
		}

		return fmt.Errorf("Error deleting project (%s): %v", id, err)
	}

	return err
}
