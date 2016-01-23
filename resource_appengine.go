package main

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"google.golang.org/api/appengine/v1beta4"
	"google.golang.org/api/googleapi"
)

func resourceAppengine() *schema.Resource {
	return &schema.Resource{
		Create: resourceAppengineCreate,
		Read:   resourceAppengineRead,
		Delete: resourceAppenginetDelete,

		Schema: map[string]*schema.Schema{
			"moduleName": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"version": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"gstorageBucket": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"gstorageKey": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"scaling": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"minIdleInstance": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default:  "1",
						},

						"maxIdleInstance": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default: "3",
						},

						"MinPendingLatency": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default: "Automatic",
						},

						"MaxPendingLatency": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default: "Automatic",
						},
					},
				},
			},
		},
	}
}

func resourceAppengineDatasetCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	datasetRef := &bigquery.DatasetReference{DatasetId: d.Get("datasetId").(string), ProjectId: config.Project}

	dataset := &bigquery.Dataset{DatasetReference: datasetRef}

	if v, ok := d.GetOk("friendlyName"); ok {
		dataset.FriendlyName = v.(string)
	}

	if v, ok := d.GetOk("description"); ok {
		dataset.Description = v.(string)
	}

	if v, ok := d.GetOk("location"); ok {
		dataset.Location = v.(string)
	}

	if v, ok := d.GetOk("defaultTableExpirationMs"); ok {
		dataset.DefaultTableExpirationMs = v.(int64)
	}

	if v, ok := d.GetOk("access"); ok {
		accessList := make([]*bigquery.DatasetAccess, 0)
		for _, access_interface := range v.([]interface{}) {
			access_parsed := &bigquery.DatasetAccess{}
			access_raw := access_interface.(map[string]interface{})
			if role, ok := access_raw["role"]; ok {
				access_parsed.Role = role.(string)
			}
			if userByEmail, ok := access_raw["userByEmail"]; ok {
				access_parsed.UserByEmail = userByEmail.(string)
			}
			if groupByEmail, ok := access_raw["groupByEmail"]; ok {
				access_parsed.GroupByEmail = groupByEmail.(string)
			}
			if domain, ok := access_raw["domain"]; ok {
				access_parsed.Domain = domain.(string)
			}
			if specialGroup, ok := access_raw["specialGroup"]; ok {
				access_parsed.SpecialGroup = specialGroup.(string)
			}
			if view, ok := access_raw["view"]; ok {
				view_raw := view.([]interface{})
				if len(view_raw) > 1 {
					fmt.Errorf("There are more then one view records in a single access record, this is not valid.")
				}
				view_parsed := &bigquery.TableReference{}
				view_zero := view_raw[0].(map[string]interface{})
				if projectId, ok := view_zero["projectId"]; ok {
					view_parsed.ProjectId = projectId.(string)
				}
				if datasetId, ok := view_zero["datasetId"]; ok {
					view_parsed.DatasetId = datasetId.(string)
				}
				if tableId, ok := view_zero["tableId"]; ok {
					view_parsed.TableId = tableId.(string)
				}
				access_parsed.View = view_parsed
			}

			accessList = append(accessList, access_parsed)
		}

		dataset.Access = accessList
	}

	call := config.clientAppengine.Datasets.Insert(config.Project, dataset)
	_, err := call.Do()
	if err != nil {
		return err
	}

	return resourceAppengineRead(d, meta)
}

func resourceAppengineRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	call := config.clientAppengine.Datasets.Get(config.Project, d.Get("datasetId").(string))
	res, err := call.Do()
	if err != nil {
		if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 404 {
			// The resource doesn't exist anymore
			d.SetId("")

			return nil
		}
		return fmt.Errorf("Failed to read bigquery dataset %s with err: %q", d.Get("datasetId").(string), err)
	}

	d.SetId(res.Id)
	d.Set("self_link", res.SelfLink)
	d.Set("lastModifiedTime", res.LastModifiedTime)
	d.Set("id", res.Id)
	return nil
}

func resourceAppengineDatasetUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceAppengineDatasetDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	call := config.clientAppengine.Datasets.Delete(config.Project, d.Get("datasetId").(string))
	err := call.Do()
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
