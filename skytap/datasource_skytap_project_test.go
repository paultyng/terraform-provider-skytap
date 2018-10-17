package skytap

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceSkytapProjectBasic(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSkytapProjectConfigBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapProjectExists("data.skytap_project.bar"),
					resource.TestCheckResourceAttr("data.skytap_project.bar", "name", fmt.Sprintf("tftest-project-data-%d", rInt)),
					resource.TestCheckResourceAttrSet("data.skytap_project.bar", "summary"),
					resource.TestCheckResourceAttr("data.skytap_project.bar", "auto_add_role_name", ""),
					resource.TestCheckResourceAttr("data.skytap_project.bar", "show_project_members", "true"),
				),
			},
		},
	})
}

func testAccDataSourceSkytapProjectConfigBasic(rInt int) string {
	return fmt.Sprintf(`
resource "skytap_project" "foo" {
	name = "tftest-project-data-%d"
	summary = "This is a project created by the skytap terraform provider acceptance test"
}

data "skytap_project" "bar" {
	name = "${skytap_project.foo.name}"
}`, rInt)
}
