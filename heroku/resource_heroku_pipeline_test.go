package heroku

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	heroku "github.com/heroku/heroku-go/v5"
)

func TestAccHerokuPipeline_Basic(t *testing.T) {
	var pipeline heroku.Pipeline
	pipelineName := fmt.Sprintf("tftest-%s", acctest.RandString(10))
	// pipelineName2 := fmt.Sprintf("%s-2", pipelineName)
	ownerID := testAccConfig.GetUserIDOrSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckHerokuPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuPipeline_basic(pipelineName, ownerID, "team"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHerokuPipelineExists("heroku_pipeline.foobar", &pipeline),
					// resource.TestCheckResourceAttr(
					// 	"heroku_pipeline.foobar", "name", pipelineName),
					// resource.TestCheckResourceAttr(
					// 	"heroku_pipeline.foobar", "owner.0.id", ownerID),
				),
			},
			// {
			// 	Config: testAccCheckHerokuPipeline_basic(pipelineName2, ownerID, "team"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr(
			// 			"heroku_pipeline.foobar", "name", pipelineName2),
			// 		resource.TestCheckResourceAttr(
			// 			"heroku_pipeline.foobar", "owner.0.id", ownerID),
			// 	),
			// },
		},
	})
}

// func TestAccHerokuPipeline_NoOwner(t *testing.T) {
// 	var pipeline heroku.Pipeline
// 	pipelineName := fmt.Sprintf("tftest-%s", acctest.RandString(10))
// 	ownerID := testAccConfig.GetUserIDOrSkip(t)

// 	resource.Test(t, resource.TestCase{
// 		PreCheck:     func() { testAccPreCheck(t) },
// 		Providers:    testAccProviders,
// 		CheckDestroy: testAccCheckHerokuPipelineDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccCheckHerokuPipeline_NoOwner(pipelineName),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckHerokuPipelineExists("heroku_pipeline.foobar", &pipeline),
// 					resource.TestCheckResourceAttr(
// 						"heroku_pipeline.foobar", "name", pipelineName),
// 					resource.TestCheckResourceAttr(
// 						"heroku_pipeline.foobar", "owner.0.id", ownerID),
// 				),
// 			},
// 		},
// 	})
// }

// func TestAccHerokuPipeline_InvalidOwnerID(t *testing.T) {
// 	pipelineName := fmt.Sprintf("tftest-%s", acctest.RandString(10))

// 	resource.Test(t, resource.TestCase{
// 		PreCheck:     func() { testAccPreCheck(t) },
// 		Providers:    testAccProviders,
// 		CheckDestroy: testAccCheckHerokuPipelineDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config:      testAccCheckHerokuPipeline_basic(pipelineName, "im-an-invalid-owner-id", "user"),
// 				ExpectError: regexp.MustCompile(`expected "owner.0.id" to be a valid UUID`),
// 			},
// 		},
// 	})
// }

// func TestAccHerokuPipeline_InvalidOwnerType(t *testing.T) {
// 	pipelineName := fmt.Sprintf("tftest-%s", acctest.RandString(10))

// 	resource.Test(t, resource.TestCase{
// 		PreCheck:     func() { testAccPreCheck(t) },
// 		Providers:    testAccProviders,
// 		CheckDestroy: testAccCheckHerokuPipelineDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config:      testAccCheckHerokuPipeline_basic(pipelineName, "16d1c25f-d879-4f4d-ad1b-d807169aaa1c", "invalid"), // not real UUID
// 				ExpectError: regexp.MustCompile(`expected owner.0.type to be one of \[team user], got invalid`),
// 			},
// 		},
// 	})
// }

func testAccCheckHerokuPipeline_basic(pipelineName, pipelineOwnerID, pipelineOwnerType string) string {
	return fmt.Sprintf(`
resource "heroku_pipeline" "foobar" {
  name = "%s"
  owner {
	id = "%s"
	type = "%s"
  }
}
`, pipelineName, pipelineOwnerID, pipelineOwnerType)
}

func testAccCheckHerokuPipeline_NoOwner(pipelineName string) string {
	return fmt.Sprintf(`
resource "heroku_pipeline" "foobar" {
  name = "%s"
}
`, pipelineName)
}

func testAccCheckHerokuPipelineExists(n string, pipeline *heroku.Pipeline) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		log.Printf("[RES STATE] %q", rs)

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		log.Printf("[RES] found")

		if rs.Primary.ID == "" {
			return fmt.Errorf("No pipeline name set")
		}

		log.Printf("[PRIMARY ID] %s", rs.Primary.ID)

		client := testAccProvider.Meta().(*Config).Api

		foundPipeline, err := client.PipelineInfo(context.TODO(), rs.Primary.ID)
		if err != nil {
			return err
		}

		if foundPipeline.ID != rs.Primary.ID {
			return fmt.Errorf("Pipeline not found")
		}

		*pipeline = *foundPipeline

		return nil
	}
}

func testAccCheckHerokuPipelineDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Config).Api

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "heroku_pipeline" {
			continue
		}

		_, err := client.PipelineInfo(context.TODO(), rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("Pipeline still exists")
		}
	}

	return nil
}
