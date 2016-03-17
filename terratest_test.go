// Integration tests that test cross-package functionality in AWS.
package terratest

import (
	"path"
	"testing"

	"github.com/gruntwork-io/terratest/aws"
	"github.com/gruntwork-io/terratest/util"
	"fmt"
"strings"
)

// This is the directory where our test fixtures are.
const fixtureDir = "./test-fixtures"

func TestUploadKeyPair(t *testing.T) {
	t.Parallel()

	// Assign randomly generated values
	region := aws.GetRandomRegion()
	id := util.UniqueId()

	// Create the keypair
	keyPair, err := util.GenerateRSAKeyPair(2048)
	if err != nil {
		t.Errorf("Failed to generate keypair: %s\n", err.Error())
	}

	// Create key in EC2
	t.Logf("Creating EC2 Keypair %s in %s...", id, region)
	defer aws.DeleteEC2KeyPair(region, id)
	aws.CreateEC2KeyPair(region, id, keyPair.PublicKey)
}

func TestTerraformApplyOnMinimalExample(t *testing.T) {
	t.Parallel()

	rand, err := CreateRandomResourceCollection()
	defer rand.DestroyResources()
	if err != nil {
		t.Errorf("Failed to create random resource collection: %s\n", err.Error())
	}

	vars := make(map[string]string)
	vars["aws_region"] = rand.AwsRegion
	vars["ec2_key_name"] = rand.KeyPair.Name
	vars["ec2_instance_name"] = rand.UniqueId
	vars["ec2_image"] = rand.AmiId

	ao := NewApplyOptions()
	ao.TestName = "Test - TestTerraformApplyMainFunction"
	ao.TemplatePath = path.Join(fixtureDir, "minimal-example")
	ao.Vars = vars
	ao.AttemptTerraformRetry = false

	_, err = ApplyAndDestroy(ao)
	if err != nil {
		t.Fatalf("Failed to ApplyAndDestroy: %s", err.Error())
	}
}

func TestTerraformApplyOnMinimalExampleWithRetry(t *testing.T) {
	t.Parallel()

	rand, err := CreateRandomResourceCollection()
	defer rand.DestroyResources()
	if err != nil {
		t.Errorf("Failed to create random resource collection: %s\n", err.Error())
	}

	vars := make(map[string]string)
	vars["aws_region"] = rand.AwsRegion
	vars["ec2_key_name"] = rand.KeyPair.Name
	vars["ec2_instance_name"] = rand.UniqueId
	vars["ec2_image"] = rand.AmiId

	ao := NewApplyOptions()
	ao.TestName = "Test - TestTerraformApplyMainFunction"
	ao.TemplatePath = path.Join(fixtureDir, "minimal-example")
	ao.Vars = vars
	ao.AttemptTerraformRetry = true

	_, err = ApplyAndDestroy(ao)
	if err != nil {
		t.Fatalf("Failed to ApplyAndDestroy: %s", err.Error())
	}
}

func TestApplyOrDestroyFailsOnTerraformError(t *testing.T) {
	t.Parallel()

	rand, err := CreateRandomResourceCollection()
	defer rand.DestroyResources()
	if err != nil {
		t.Errorf("Failed to create random resource collection: %s\n", err.Error())
	}

	vars := make(map[string]string)
	vars["aws_region"] = rand.AwsRegion
	vars["ec2_key_name"] = rand.KeyPair.Name
	vars["ec2_instance_name"] = rand.UniqueId
	vars["ec2_image"] = rand.AmiId

	ao := NewApplyOptions()
	ao.TestName = "Test - TestTerraformApplyMainFunction"
	ao.TemplatePath = path.Join(fixtureDir, "minimal-example-with-error")
	ao.Vars = vars
	ao.AttemptTerraformRetry = true

	_, err = ApplyAndDestroy(ao)
	if err != nil {
		fmt.Printf("Received expected failure message: %s. Continuing on...", err.Error())
	} else {
		t.Fatalf("Expected a terraform apply error but ApplyAndDestroy did not return an error.")
	}
}

// Test that ApplyAndDestroy correctly retries a terraform apply when a "retryableErrorMessage" is detected. We validate
// this by scanning for a string in the output that explicitly indicates a terraform apply retry.
func TestTerraformApplyOnMinimalExampleWithRetryableErrorMessages(t *testing.T) {
	t.Parallel()

	rand, err := CreateRandomResourceCollection()
	defer rand.DestroyResources()
	if err != nil {
		t.Errorf("Failed to create random resource collection: %s\n", err.Error())
	}

	vars := make(map[string]string)
	vars["aws_region"] = rand.AwsRegion
	vars["ec2_key_name"] = rand.KeyPair.Name
	vars["ec2_instance_name"] = rand.UniqueId
	vars["ec2_image"] = rand.AmiId

	ao := NewApplyOptions()
	ao.TestName = "Test - TestTerraformApplyMainFunction"
	ao.TemplatePath = path.Join(fixtureDir, "minimal-example-with-error")
	ao.Vars = vars
	ao.AttemptTerraformRetry = true
	ao.RetryableTerraformErrors = make(map[string]string)
	ao.RetryableTerraformErrors["aws_instance.demo: Error launching source instance: InvalidKeyPair.NotFound"] = "This error was deliberately added to the template."

	output, err := ApplyAndDestroy(ao)
	if err != nil {
		if strings.Contains(output, "**TERRAFORM-RETRY**") {
			fmt.Println("Expected error was caught and a retry was attempted.")
		} else {
			t.Fatalf("Failed to catch expected error: %s", err.Error())
		}
	} else {
		t.Fatalf("Expected this template to have an error, but no error was thrown.")
	}

}

// Test that ApplyAndDestroy correctly avoids a retry when no "retryableErrorMessage" is detected.
func TestTerraformApplyOnMinimalExampleWithRetryableErrorMessagesDoesNotRetry(t *testing.T) {
	t.Parallel()

	rand, err := CreateRandomResourceCollection()
	defer rand.DestroyResources()
	if err != nil {
		t.Errorf("Failed to create random resource collection: %s\n", err.Error())
	}

	vars := make(map[string]string)
	vars["aws_region"] = rand.AwsRegion
	vars["ec2_key_name"] = rand.KeyPair.Name
	vars["ec2_instance_name"] = rand.UniqueId
	vars["ec2_image"] = rand.AmiId

	ao := NewApplyOptions()
	ao.TestName = "Test - TestTerraformApplyMainFunction"
	ao.TemplatePath = path.Join(fixtureDir, "minimal-example-with-error")
	ao.Vars = vars
	ao.AttemptTerraformRetry = true
	ao.RetryableTerraformErrors = make(map[string]string)
	ao.RetryableTerraformErrors["I'm a message that shouldn't show up in the output"] = ""

	output, err := ApplyAndDestroy(ao)
	if err != nil {
		if strings.Contains(output, "**TERRAFORM-RETRY**") {
			t.Fatalf("Expected no terraform retry but instead a retry was attempted.")
		} else {
			fmt.Println("An error occurred and a retry was correctly avoided.")
		}
	} else {
		t.Fatalf("Expected this template to have an error, but no error was thrown.")
	}
}
