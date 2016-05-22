package main

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAppengineCreate(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppengineDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAppengine,
				Check: resource.ComposeTestCheckFunc(
					testAccAppengineExists("googleappengine_app.foobar"),
				),
			},
		},
	})
}

func TestAccPythonAppengineCreate(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppengineDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAppenginePython,
				Check: resource.ComposeTestCheckFunc(
					testAccAppengineExists("googleappengine_app.foobar"),
				),
			},
		},
	})
}

func testAccCheckAppengineDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "googleappengine_app" {
			continue
		}

		config := testAccProvider.Meta().(*Config)
		_, err := config.clientAppengine.Apps.Modules.Versions.Get(config.Project, rs.Primary.Attributes["moduleName"], rs.Primary.Attributes["version"]).Do()
		if err != nil {
			fmt.Errorf("Application still present")
		}
	}

	return nil
}

func testAccAppengineExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}
		config := testAccProvider.Meta().(*Config)
		_, err := config.clientAppengine.Apps.Modules.Versions.Get(config.Project, rs.Primary.Attributes["moduleName"], rs.Primary.Attributes["version"]).Do()
		if err != nil {
			fmt.Errorf("Application not present")
		}

		return nil
	}
}

const testAccAppengine = `
resource "googleappengine_app" "foobar" {
	moduleName = "foobar"
	version = "foobaz"
	gstorageBucket = "build-artifacts-public-eu"
	gstorageKey = "hxtest-1.0-SNAPSHOT/"
	runtime = "java7"
	
	scaling {
		minIdleInstances = 1
		maxIdleInstances = 3
		minPendingLatency = "1s"
		maxPendingLatency = "10s"
	}
	
	topicName = "projects/hx-test/topics/notarealtopic"
}`

const testAccAppenginePython = `
resource "googleappengine_app" "foobar" {
	moduleName = "foobar"
	version = "foobaz"
	gstorageBucket = "build-artifacts-public-eu"
	gstorageKey = "python-test-app/"
	runtime = "python27"
	scriptName = "guestbook.app"
	pythonUrlRegex = "/.*"
	
	scaling {
		minIdleInstances = 1
		maxIdleInstances = 3
		minPendingLatency = "1s"
		maxPendingLatency = "10s"
	}
	
	topicName = "projects/hx-test/topics/notarealtopic"
}`
