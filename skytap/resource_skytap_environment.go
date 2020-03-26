package skytap

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/skytap/skytap-sdk-go/skytap"
	"github.com/terraform-providers/terraform-provider-skytap/skytap/utils"
)

func resourceSkytapEnvironment() *schema.Resource {
	return &schema.Resource{
		Create: resourceSkytapEnvironmentCreate,
		Read:   resourceSkytapEnvironmentRead,
		Update: resourceSkytapEnvironmentUpdate,
		Delete: resourceSkytapEnvironmentDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"template_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"description": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"outbound_traffic": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					DiffSuppressFunc: caseInsensitiveSuppress,
				},
				Set: stringCaseSensitiveHash,
			},

			"label": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"category": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"user_data": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},

			"routable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"suspend_on_idle": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(300, 86400),
			},

			"suspend_at_time": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},

			"shutdown_on_idle": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(300, 86400),
			},

			"shutdown_at_time": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},
		},
	}
}

func resourceSkytapEnvironmentCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).environmentsClient
	ctx, cancel := stopContextForCreate(d, meta.(*SkytapClient))
	defer cancel()

	templateID := d.Get("template_id").(string)
	name := d.Get("name").(string)
	outboundTraffic := d.Get("outbound_traffic").(bool)
	routable := d.Get("routable").(bool)

	opts := skytap.CreateEnvironmentRequest{
		TemplateID:      &templateID,
		Name:            &name,
		OutboundTraffic: &outboundTraffic,
		Routable:        &routable,
	}

	if v, ok := d.GetOk("description"); ok {
		opts.Description = utils.String(v.(string))
	}

	if v, ok := d.GetOk("suspend_on_idle"); ok {
		opts.SuspendOnIdle = utils.Int(v.(int))
	}

	if v, ok := d.GetOk("suspend_at_time"); ok {
		opts.SuspendAtTime = utils.String(v.(string))
	}

	if v, ok := d.GetOk("shutdown_on_idle"); ok {
		opts.ShutdownOnIdle = utils.Int(v.(int))
	}

	if v, ok := d.GetOk("shutdown_at_time"); ok {
		opts.ShutdownAtTime = utils.String(v.(string))
	}

	if tag, ok := d.GetOk("tags"); ok {
		opts.Tags = environmentCreateTags(tag.(*schema.Set))
	}

	if label, ok := d.GetOk("label"); ok {
		opts.Labels = environmentCreateLabels(label.(*schema.Set))
	}

	if v, ok := d.GetOk("user_data"); ok {
		opts.UserData = utils.String(v.(string))
	}

	log.Printf("[INFO] environment create")
	log.Printf("[TRACE] environment create options: %v", spew.Sdump(opts))
	environment, err := client.Create(ctx, &opts)
	if err != nil {
		return fmt.Errorf("error creating environment: %v", err)
	}

	if environment.ID == nil {
		return fmt.Errorf("environment ID is not set")
	}
	environmentID := *environment.ID
	d.SetId(environmentID)

	log.Printf("[INFO] environment created: %s", *environment.ID)
	log.Printf("[TRACE] environment created: %v", spew.Sdump(environment))

	stateConf := &resource.StateChangeConf{
		Pending:    environmentPendingCreateRunstates,
		Target:     environmentTargetCreateRunstates,
		Refresh:    environmentCreateRunstateRefreshFunc(ctx, d, meta),
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		MinTimeout: minTimeout * time.Second,
		Delay:      delay * time.Second,
	}

	log.Printf("[INFO] Waiting for environment (%s) to complete", d.Id())
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for environment (%s) to complete: %s", d.Id(), err)
	}

	return resourceSkytapEnvironmentRead(d, meta)
}

func resourceSkytapEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).environmentsClient
	ctx, cancel := stopContextForRead(d, meta.(*SkytapClient))
	defer cancel()

	id := d.Id()

	log.Printf("[INFO] retrieving environment: %s", id)
	environment, err := client.Get(ctx, id)
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			log.Printf("[DEBUG] environment (%s) was not found - removing from state", id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error retrieving environment (%s): %v", id, err)
	}

	var routable bool
	if environment.OutboundTraffic != nil {
		routable = *environment.OutboundTraffic
	}

	// The templateID is not set as it is used to build the environment and is not returned by the environment response.
	// If this attribute is changed, this environment will be rebuilt
	d.Set("name", environment.Name)
	d.Set("description", environment.Description)
	d.Set("outbound_traffic", environment.OutboundTraffic)
	d.Set("routable", routable)
	d.Set("suspend_on_idle", environment.SuspendOnIdle)
	d.Set("suspend_at_time", environment.SuspendAtTime)
	d.Set("shutdown_on_idle", environment.ShutdownOnIdle)
	d.Set("shutdown_at_time", environment.ShutdownAtTime)
	d.Set("user_data", environment.UserData)

	if environment.Tags != nil {
		if err = d.Set("tags", flattenTags(environment.Tags)); err != nil {
			return err
		}
	}

	if environment.LabelCount != nil && *environment.LabelCount > 0 {
		if err = d.Set("label", flattenLabels(environment.Labels)); err != nil {
			return err
		}
	} else {
		emptyLabels := make([]interface{}, 0)
		d.Set("label", emptyLabels)
	}

	log.Printf("[INFO] environment retrieved: %s", id)
	log.Printf("[TRACE] environment retrieved: %v", spew.Sdump(environment))

	return err
}

func resourceSkytapEnvironmentUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).environmentsClient
	ctx, cancel := stopContextForUpdate(d, meta.(*SkytapClient))
	defer cancel()

	id := d.Id()

	name := d.Get("name").(string)
	outboundTraffic := d.Get("outbound_traffic").(bool)
	routable := d.Get("routable").(bool)

	opts := skytap.UpdateEnvironmentRequest{
		Name:            &name,
		OutboundTraffic: &outboundTraffic,
		Routable:        &routable,
	}

	if v, ok := d.GetOk("description"); ok {
		opts.Description = utils.String(v.(string))
	}

	if v, ok := d.GetOk("suspend_on_idle"); ok {
		opts.SuspendOnIdle = utils.Int(v.(int))
	}

	if v, ok := d.GetOk("suspend_at_time"); ok {
		opts.SuspendAtTime = utils.String(v.(string))
	}

	if v, ok := d.GetOk("shutdown_on_idle"); ok {
		opts.ShutdownOnIdle = utils.Int(v.(int))
	}

	if v, ok := d.GetOk("shutdown_at_time"); ok {
		opts.ShutdownAtTime = utils.String(v.(string))
	}

	log.Printf("[INFO] environment update: %s", id)
	log.Printf("[TRACE] environment update options: %v", spew.Sdump(opts))

	d.Partial(true)
	environment, err := client.Update(ctx, id, &opts)
	if err != nil {
		return fmt.Errorf("error updating environment (%s): %v", id, err)
	}

	log.Printf("[INFO] environment updated: %s", id)
	log.Printf("[TRACE] environment updated: %v", spew.Sdump(environment))
	if err = waitForEnvironmentReady(ctx, d, meta, *environment.ID, schema.TimeoutUpdate); err != nil {
		return err
	}

	if d.HasChange("tags") {
		old, new := d.GetChange("tags")
		newTagsSet := new.(*schema.Set)
		oldTagSet := old.(*schema.Set)

		tagsToRemove := oldTagSet.Difference(newTagsSet)
		if tagsToRemove.Len() > 0 {
			// Create a dictionary to transform tags in ids
			tagDictionary := make(map[string]string)
			for _, t := range environment.Tags {
				tagDictionary[*t.Value] = *t.ID
			}
			// No batch removal supported by the api, remove one by one
			for _, t := range tagsToRemove.List() {
				if err := client.DeleteTag(ctx, *environment.ID, tagDictionary[t.(string)]); err != nil {
					return err
				}
			}
		}

		tagsToAdd := newTagsSet.Difference(oldTagSet)
		if err := client.CreateTags(ctx, *environment.ID, environmentCreateTags(tagsToAdd)); err != nil {
			return err
		}

		d.SetPartial("tags")
	}

	if d.HasChange("label") {
		old, new := d.GetChange("label")
		remove := old.(*schema.Set).Difference(new.(*schema.Set))
		add := new.(*schema.Set).Difference(old.(*schema.Set))

		for _, l := range remove.List() {
			label := l.(map[string]interface{})
			if err = client.DeleteLabel(ctx, *environment.ID, label["id"].(string)); err != nil {
				return err
			}
		}
		if err = client.CreateLabels(ctx, *environment.ID, environmentCreateLabels(add)); err != nil {
			return err
		}
		d.SetPartial("label")
	}

	if d.HasChange("user_data") {
		if err := client.UpdateUserData(ctx, *environment.ID, utils.String(d.Get("user_data").(string))); err != nil {
			return err
		}
		d.SetPartial("user_data")
	}

	d.Partial(false)
	return resourceSkytapEnvironmentRead(d, meta)
}

func waitForEnvironmentReady(ctx context.Context, d *schema.ResourceData, meta interface{}, environmentID string, schemaTimeout string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    environmentPendingUpdateRunstates,
		Target:     environmentTargetUpdateRunstates,
		Refresh:    environmentUpdateRunstateRefreshFunc(ctx, meta, environmentID),
		Timeout:    d.Timeout(schemaTimeout),
		MinTimeout: minTimeout * time.Second,
		Delay:      delay * time.Second,
	}

	log.Printf("[INFO] Waiting for environment (%s) to complete", environmentID)
	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for environment (%s) to complete: %s", environmentID, err)
	}
	return nil
}

func resourceSkytapEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*SkytapClient).environmentsClient
	ctx, cancel := stopContextForDelete(d, meta.(*SkytapClient))
	defer cancel()

	id := d.Id()

	log.Printf("[INFO] destroying environment: %s", id)
	err := client.Delete(ctx, id)
	if err != nil {
		if utils.ResponseErrorIsNotFound(err) {
			log.Printf("[DEBUG] environment (%s) was not found - assuming removed", id)
			return nil
		}

		return fmt.Errorf("error deleting environment (%s): %v", id, err)
	}

	log.Printf("[INFO] environment destroyed: %s", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"false"},
		Target:     []string{"true"},
		Refresh:    environmentDeleteRefreshFunc(ctx, d, meta),
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		MinTimeout: minTimeout * time.Second,
		Delay:      delay * time.Second,
	}

	log.Printf("[INFO] Waiting for environment (%s) to complete", d.Id())
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for environment (%s) to complete: %s", d.Id(), err)
	}

	return err
}

var environmentPendingCreateRunstates = []string{
	string(skytap.EnvironmentRunstateBusy),
}

var environmentTargetCreateRunstates = []string{
	string(skytap.EnvironmentRunstateRunning),
}

var environmentPendingUpdateRunstates = []string{
	string(skytap.EnvironmentRunstateBusy),
}

var environmentTargetUpdateRunstates = []string{
	string(skytap.EnvironmentRunstateRunning),
	string(skytap.EnvironmentRunstateStopped),
	string(skytap.EnvironmentRunstateSuspended),
}

func environmentCreateRunstateRefreshFunc(
	ctx context.Context, d *schema.ResourceData, meta interface{}) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		client := meta.(*SkytapClient).environmentsClient

		id := d.Id()

		log.Printf("[DEBUG] retrieving environment: %s", id)
		environment, err := client.Get(ctx, id)

		if err != nil {
			return nil, "", fmt.Errorf("error retrieving environment (%s) when waiting: %v", id, err)
		}

		computedRunstate := skytap.EnvironmentRunstateRunning
		for i := 0; i < *environment.VMCount; i++ {
			if *environment.VMs[i].Runstate != skytap.VMRunstateRunning {
				computedRunstate = skytap.EnvironmentRunstateBusy
				break
			}
		}

		log.Printf("[DEBUG] environment status (%s): %s", id, *environment.Runstate)

		return environment, string(computedRunstate), nil
	}
}

func environmentUpdateRunstateRefreshFunc(ctx context.Context, meta interface{}, environmentID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		client := meta.(*SkytapClient).environmentsClient

		log.Printf("[DEBUG] retrieving environment: %s", environmentID)
		environment, err := client.Get(ctx, environmentID)

		if err != nil {
			return nil, "", fmt.Errorf("error retrieving environment (%s) when waiting: %v", environmentID, err)
		}

		log.Printf("[DEBUG] environment (%s): %s", environmentID, *environment.Runstate)

		return environment, string(*environment.Runstate), nil
	}
}

func environmentDeleteRefreshFunc(
	ctx context.Context, d *schema.ResourceData, meta interface{}) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		client := meta.(*SkytapClient).environmentsClient

		id := d.Id()

		log.Printf("[DEBUG] retrieving environment: %s", id)
		environment, err := client.Get(ctx, id)

		var removed = "false"
		if err != nil {
			if utils.ResponseErrorIsNotFound(err) {
				log.Printf("[DEBUG] environment (%s) has been removed.", id)
				removed = "true"
			} else {
				return nil, "", fmt.Errorf("error retrieving environment (%s) when waiting: %v", id, err)
			}
		}

		return environment, removed, nil
	}
}

func environmentCreateTags(vs *schema.Set) []*skytap.CreateTagRequest {
	createTagRequests := make([]*skytap.CreateTagRequest, vs.Len())
	for i, v := range vs.List() {
		createTagRequests[i] = &skytap.CreateTagRequest{
			Tag: v.(string),
		}
	}
	return createTagRequests
}

func environmentCreateLabels(vs *schema.Set) []*skytap.CreateLabelRequest {
	createLabelsRequest := make([]*skytap.CreateLabelRequest, vs.Len())
	for i, v := range vs.List() {
		elem := v.(map[string]interface{})
		createLabelsRequest[i] = &skytap.CreateLabelRequest{
			Category: utils.String(elem["category"].(string)),
			Value:    utils.String(elem["value"].(string)),
		}
	}
	return createLabelsRequest
}
