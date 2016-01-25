package main

import (
	"fmt"
	"log"
	"time"
	"strings"
	"github.com/hashicorp/terraform/helper/schema"
	"google.golang.org/api/appengine/v1beta4"
	"google.golang.org/api/storage/v1"
)

func resourceAppengine() *schema.Resource {
	return &schema.Resource{
		Create: resourceAppengineCreate,
		Read:   resourceAppengineRead,
		Delete: resourceAppengineDelete,

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
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"minIdleInstances": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Default:  "1",
						},

						"maxIdleInstances": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Default: "3",
						},

						"minPendingLatency": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default: "Automatic",
						},

						"maxPendingLatency": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default: "Automatic",
						},
					},
				},
			},
			"servingStatus": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

var (
	remoteBase = "https://storage.googleapis.com/"
)


func urlHandlers() ([]*appengine.UrlMap) {
	handlers := make([]*appengine.UrlMap, 0)
		handlers = append(handlers, &appengine.UrlMap{
			SecurityLevel: "SECURE_OPTIONAL",
			Login: "LOGIN_OPTIONAL",
			UrlRegex:"/", 
			Script:&appengine.ScriptHandler{
				ScriptPath:"unused",
			},
		})
		handlers = append(handlers, &appengine.UrlMap{
			SecurityLevel: "SECURE_OPTIONAL",
			Login: "LOGIN_OPTIONAL",
			UrlRegex:"/.*/", 
			Script:&appengine.ScriptHandler{
				ScriptPath:"unused",
			},
		})
		handlers = append(handlers, &appengine.UrlMap{
			SecurityLevel: "SECURE_OPTIONAL",
			Login: "LOGIN_OPTIONAL",
			UrlRegex:"/_ah/.*", 
			Script:&appengine.ScriptHandler{
				ScriptPath:"unused",
			},
		})
		handlers = append(handlers, &appengine.UrlMap{
			SecurityLevel: "SECURE_OPTIONAL",
			Login: "LOGIN_OPTIONAL",
			UrlRegex:"/endpoint", 
			Script:&appengine.ScriptHandler{
				ScriptPath:"unused",
			},
		})
		
		return handlers
}


// known issues with this function:
//   assumes "/" is delimiter in gstorage and forces that to be last char in key
//   only searches first page, if more then 1k files to load, will only grab first 1k
func generateFileList(d *schema.ResourceData, config *Config) (map[string]appengine.FileInfo, error) {
	listService := storage.NewObjectsService(config.clientStorage)
	bucket := d.Get("gstorageBucket").(string)
	listCall := listService.List(bucket)
	key := d.Get("gstorageKey").(string)
	lastChar := key[len(key)-1:]
	if lastChar != "/" {
		key = key + "/"
	}
	listCall = listCall.Prefix(key)
	objs, err := listCall.Do()
	if err != nil {
		return nil, err
	}
	
	files := make(map[string]appengine.FileInfo)
	for _, obj := range objs.Items {
		onDiskName := strings.Replace(obj.Name, key, "", 1)  // trims key from file name
		inCloudURL := remoteBase + bucket + "/" + obj.Name
		files[onDiskName] = appengine.FileInfo{SourceUrl:inCloudURL} 
	}
	
	return files, nil
}

func resourceAppengineCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)


	/*scaling_raw := d.Get("scaling").([]interface{})
	if len(scaling_raw) > 1 {
		return fmt.Errorf("User supplied more then one scaling setting.  This is wrong")
	}
	
	scale := scaling_raw[0].(map[string]interface{})
	automaticScaling := &appengine.AutomaticScaling{
		MinIdleInstances: int64(scale["minIdleInstances"].(int)),
		MaxIdleInstances: int64(scale["maxIdleInstances"].(int)),
		MinPendingLatency: scale["minPendingLatency"].(string),
		MaxPendingLatency: scale["maxPendingLatency"].(string),
	}*/
	
	files, err := generateFileList(d, config)
	if err != nil {
		return err
	}
	deployment := &appengine.Deployment{Files:files}
	
	handlers := urlHandlers()
	
	inbound_services := make([]string, 1)
	inbound_services[0] = "INBOUND_SERVICE_WARMUP"
	
	//  Version object for this module 
	version := &appengine.Version{
		//AutomaticScaling: automaticScaling, 
		Deployment:deployment, 
		Handlers: handlers, 
		Id: d.Get("version").(string), 
		Runtime: "java7",
		//InstanceClass: "F2",  this is exploding.  not sure why
		InboundServices: inbound_services,
		Threadsafe: true,
	}
	
	//  create the application
	moduleService := appengine.NewAppsModulesVersionsService(config.clientAppengine)
	createCall := moduleService.Create(config.Project, d.Get("moduleName").(string), version)
	operation, err := createCall.Do()
	if err != nil {
		return err
	}
	
	err = operationWait(operation, config)
	if err != nil {
		return err
	}
	
	return resourceAppengineRead(d, meta)
}

func operationWait(operation *appengine.Operation, config *Config) (error) {
	//  wait for the creation to complete
	operationService := appengine.NewAppsOperationsService(config.clientAppengine)
	operationGet := operationService.Get(config.Project, strings.Replace(operation.Name, "apps/"+config.Project+"/operations/", "", 1))
	carryon := true
	for carryon {
		operation, err := operationGet.Do()
		if err != nil {
			return err
		}
		carryon = !operation.Done
		time.Sleep(10*time.Second)
	}
	
	//   if it failed, explode
	if operation.Error != nil {
		log.Printf("[DEBUG] status list from bad operation: %q", operation.Error.Details)
		return fmt.Errorf(operation.Error.Message)
	}
	
	return nil
}

func resourceAppengineRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	moduleService := appengine.NewAppsModulesVersionsService(config.clientAppengine)
	getCall := moduleService.Get(config.Project, d.Get("moduleName").(string), d.Get("version").(string))
	version, err := getCall.Do()
	if err != nil {
		return err
	}

	d.SetId(version.Name)
	d.Set("servingStatus", version.ServingStatus)
	return nil
}

func resourceAppengineDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	moduleService := appengine.NewAppsModulesVersionsService(config.clientAppengine)
	deleteCall := moduleService.Delete(config.Project, d.Get("moduleName").(string), d.Get("version").(string))
	operation, err := deleteCall.Do()
	if err != nil {
		return err
	}

	err =  operationWait(operation, config)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
