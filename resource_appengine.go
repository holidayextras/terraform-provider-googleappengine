package main

import (
	"fmt"
	"log"
	"time"
	"regexp"
	"strings"
	"strconv"
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
			"threadsafe": &schema.Schema{
				Type: 	  schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  true,
			},
			
			"runtime": &schema.Schema{
				Type:	  schema.TypeString,
				Required: true,
				ForceNew: true,	
			},
			"resource_version": &schema.Schema{
				Type:     schema.TypeList,
				ForceNew: true,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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
			"topicName": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"scriptName": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"pythonUrlRegex": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"env_args": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     schema.TypeString,
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

func pythonUrlHandlers(d *schema.ResourceData) ([]*appengine.UrlMap) {
	handlers := make([]*appengine.UrlMap, 0)
	handlers = append(handlers, &appengine.UrlMap{
		SecurityLevel: "SECURE_OPTIONAL",
		Login: "LOGIN_OPTIONAL",
		UrlRegex:d.Get("pythonUrlRegex").(string), 
		Script:&appengine.ScriptHandler{
			ScriptPath:d.Get("scriptName").(string),
		},
	})
		
	return handlers
}

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
	
	//  when NextPageToken is empty string, on last page
	for objs.NextPageToken != "" {
		files = objsToFilelist(objs, files, key, bucket)
		listCall.PageToken(objs.NextPageToken)
		objs, err = listCall.Do()
		if err != nil {
			return nil, err
		}
	}
	
	//  read contents of last page
	files = objsToFilelist(objs, files, key, bucket)
	
	return files, nil
}

func objsToFilelist(objs *storage.Objects, files map[string]appengine.FileInfo, key, bucket string) (map[string]appengine.FileInfo) {
	for _, obj := range objs.Items {
		if matched, _ := regexp.MatchString("[(~]", obj.Name); !matched {  // both ( and ~ are illegal file name chars
			onDiskName := strings.Replace(obj.Name, key, "", 1)  // trims key from file name
			inCloudURL := remoteBase + bucket + "/" + obj.Name
			files[onDiskName] = appengine.FileInfo{SourceUrl:inCloudURL}
		}
	}
	return files
}

func validateLatency(latency string) (string, error) {
	lastChar := latency[len(latency)-1:]
	if lastChar != "s" {
		return "", fmt.Errorf("latency values must be between 1 and 15 seconds in the form: 3s")
	}
	latency_i, err := strconv.Atoi(latency[:len(latency)-1])
	if err != nil {
		return "", err
	}
	if latency_i < 1 || latency_i > 15 {
		return "", fmt.Errorf("latency values must be between 1 and 15 seconds in the form: 3s")
	}
	
	return latency, nil
}

func cleanAdditionalArgs(optional_args map[string]interface{}) map[string]string {
	cleaned_opts := make(map[string]string)
	for k,v := range  optional_args {
		cleaned_opts[k] = v.(string)
	}
	return cleaned_opts
}

func resourceAppengineCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)


	scaling_raw := d.Get("scaling").([]interface{})
	if len(scaling_raw) > 1 {
		return fmt.Errorf("User supplied more then one scaling setting.  This is wrong")
	}
	
	
	scale := scaling_raw[0].(map[string]interface{})
	minPendingLatency, err := validateLatency(scale["minPendingLatency"].(string))
	if err != nil {
		return err
	}
	
	maxPendingLatency, err := validateLatency(scale["maxPendingLatency"].(string))
	if err != nil {
		return err
	}
	automaticScaling := &appengine.AutomaticScaling{
		MinIdleInstances: int64(scale["minIdleInstances"].(int)),
		MaxIdleInstances: int64(scale["maxIdleInstances"].(int)),
		MinPendingLatency: minPendingLatency,
		MaxPendingLatency: maxPendingLatency,
	}
		
	files, err := generateFileList(d, config)
	if err != nil {
		return err
	}
	deployment := &appengine.Deployment{Files:files}
	
	
	inbound_services := make([]string, 1)
	inbound_services[0] = "INBOUND_SERVICE_WARMUP"
	
	env_args := cleanAdditionalArgs(d.Get("env_args").(map[string]interface{}))
	env_vars := make(map[string]string,2)
	// the specific env vars in the config are given precendence so use the general map first
	for k, v := range env_args {
		env_vars[k] = v
	}
	if outputpubsub, ok := d.GetOk("topicName"); ok {
		env_vars["OUTPUTPUBSUB"] = outputpubsub.(string)
	} 
	env_vars["RETURNMESSAGEIDS"] = "true"
	
	//  Version object for this module 
	version := &appengine.Version{
		AutomaticScaling: automaticScaling, 
		Deployment:deployment, 
		Id: d.Get("version").(string), 
		Runtime: d.Get("runtime").(string),
		//InstanceClass: "F2",  this is exploding.  not sure why
		InboundServices: inbound_services,
		EnvVariables: env_vars,
		Threadsafe: d.Get("threadsafe").(bool),
	}

	if d.Get("runtime").(string) == "java7" {
		version.Handlers = urlHandlers()
	} else {
		version.Handlers = pythonUrlHandlers(d)
	}
	
	//  create the application
	moduleVersionService := appengine.NewAppsModulesVersionsService(config.clientAppengine)
	createCall := moduleVersionService.Create(config.Project, d.Get("moduleName").(string), version)
	operation, err := createCall.Do()
	if err != nil {
		return fmt.Errorf("failed to create mod ver: " + err.Error())
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
	op := &appengine.Operation{}
	var err error
	for carryon {
		op, err = operationGet.Do()
		if err != nil {
			return fmt.Errorf("Wait operation has exploded: " + err.Error())
		}
		carryon = !op.Done
		time.Sleep(10*time.Second)
		log.Printf("[DEBUG] here's the whole op: %+v", op)
	}
	

	
	//   if it failed, explode
	if op.Error != nil {
		log.Printf("[DEBUG] status list from bad operation: %q", op.Error.Details)
		return fmt.Errorf("The operation has completed with errors: " + op.Error.Message)
	}
	
	return nil
}

func resourceAppengineRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	moduleVersionService := appengine.NewAppsModulesVersionsService(config.clientAppengine)
	getCall := moduleVersionService.Get(config.Project, d.Get("moduleName").(string), d.Get("version").(string))
	version, err := getCall.Do()
	if err != nil {
		return fmt.Errorf("Couldn't find the resource: " + err.Error())
	}

	d.SetId(version.Name)
	d.Set("servingStatus", version.ServingStatus)
	return nil
}

func resourceAppengineDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	moduleVersionService := appengine.NewAppsModulesVersionsService(config.clientAppengine)
	deleteCall := moduleVersionService.Delete(config.Project, d.Get("moduleName").(string), d.Get("version").(string))
	operation, err := deleteCall.Do()
	if err != nil {
		if strings.Contains(err.Error(), "Cannot delete the final version of a service (module)") {
			moduleService := appengine.NewAppsModulesService(config.clientAppengine)
			moduleDelete := moduleService.Delete(config.Project, d.Get("moduleName").(string))
			operation, err = moduleDelete.Do()
			if err != nil {
				return err
			}
			
			err = operationWait(operation, config)
			if err != nil {
				return err
			}		
		} else {
			return err
		}
	} else {
		err = operationWait(operation, config)
		if err != nil {
			return err
		}
	}

	d.SetId("")
	return nil
}
