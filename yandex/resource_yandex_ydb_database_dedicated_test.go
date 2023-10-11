package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/ydb/v1"
)

const ydbDatabaseDedicatedResource = "yandex_ydb_database_dedicated.test-ydb-database-dedicated"

func init() {
	resource.AddTestSweepers("yandex_ydb_database_dedicated", &resource.Sweeper{
		Name: "yandex_ydb_database_dedicated",
		F:    testSweepYDBDatabaseDedicated,
	})
}

func testSweepYDBDatabaseDedicated(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.sdk.YDB().Database().List(conf.Context(), &ydb.ListDatabasesRequest{
		FolderId: conf.FolderID,
		PageSize: 1000,
	})
	if err != nil {
		return fmt.Errorf("error getting YDB databases: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Databases {
		// skip serverless, it's swept in a separate sweeper
		if _, ok := c.DatabaseType.(*ydb.Database_ServerlessDatabase); !ok {
			if !sweepYDBDatabaseDedicated(conf, c.Id) {
				result = multierror.Append(result, fmt.Errorf("failed to sweep YDB dedicated database %q", c.Id))
			}
		}
	}

	return result.ErrorOrNil()
}

func sweepYDBDatabaseDedicated(conf *Config, id string) bool {
	return sweepWithRetry(sweepYDBDatabaseDedicatedOnce, conf, "YDB dedicated database", id)
}

func unsetYDBDatabaseDeletionProtection(
	conf *Config,
	ctx context.Context,
	db *ydb.Database,
) error {
	req := &ydb.UpdateDatabaseRequest{
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{
				"deletion_protection",
			},
		},
		DatabaseId: db.Id,
	}
	req.DeletionProtection = false

	op, err := conf.sdk.YDB().Database().Update(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to unset deletion_protection for database %q: %s", db.Id, err)
	}

	sdkop, err := conf.sdk.WrapOperation(op, err)
	if err != nil {
		return err
	}

	err = sdkop.Wait(ctx)
	if err != nil {
		return err
	}
	_, err = sdkop.Response()
	return err
}

func checkAndUnsetYDBDeletionProtection(
	conf *Config,
	ctx context.Context,
	id string,
) error {
	db, err := conf.sdk.YDB().Database().Get(ctx, &ydb.GetDatabaseRequest{
		DatabaseId: id,
	})
	if err != nil {
		return fmt.Errorf("failed to get database with id %q: %s", id, err)
	}

	if !db.DeletionProtection {
		return nil
	}

	return unsetYDBDatabaseDeletionProtection(conf, ctx, db)
}

func sweepYDBDatabaseDedicatedOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexYDBDedicatedDefaultTimeout)
	defer cancel()

	err := checkAndUnsetYDBDeletionProtection(conf, ctx, id)
	if err != nil {
		return err
	}

	op, err := conf.sdk.YDB().Database().Delete(ctx, &ydb.DeleteDatabaseRequest{
		DatabaseId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func TestAccYandexYDBDatabaseDedicated_basic(t *testing.T) {
	var database ydb.Database
	databaseName := acctest.RandomWithPrefix("tf-ydb-database-dedicated")
	databaseDesc := acctest.RandomWithPrefix("tf-ydb-database-dedicated-desc")
	labelKey := acctest.RandomWithPrefix("tf-ydb-database-dedicated-label")
	labelValue := acctest.RandomWithPrefix("tf-ydb-database-dedicated-label-value")
	deletionProtection := "false"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexYDBDatabaseDedicatedDestroy,
		Steps: []resource.TestStep{
			basicYandexYDBDatabaseDedicatedTestStep(databaseName, databaseDesc, deletionProtection, labelKey, labelValue, &database),
		},
	})
}

func TestAccYandexYDBDatabaseDedicated_update(t *testing.T) {
	var database ydb.Database
	databaseName := acctest.RandomWithPrefix("tf-ydb-database-dedicated")
	databaseDesc := acctest.RandomWithPrefix("tf-ydb-database-dedicated-desc")
	labelKey := acctest.RandomWithPrefix("tf-ydb-database-dedicated-label")
	labelValue := acctest.RandomWithPrefix("tf-ydb-database-dedicated-label-value")
	deletionProtection := "true"

	databaseNameUpdated := acctest.RandomWithPrefix("tf-ydb-database-dedicated-updated")
	databaseDescUpdated := acctest.RandomWithPrefix("tf-ydb-database-dedicated-desc-updated")
	labelKeyUpdated := acctest.RandomWithPrefix("tf-ydb-database-dedicated-label-updated")
	labelValueUpdated := acctest.RandomWithPrefix("tf-ydb-database-dedicated-label-value-updated")
	deletionProtectionUpdated := "false"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexYDBDatabaseDedicatedDestroy,
		Steps: []resource.TestStep{
			basicYandexYDBDatabaseDedicatedTestStep(databaseName, databaseDesc, deletionProtection, labelKey, labelValue, &database),
			basicYandexYDBDatabaseDedicatedTestStep(databaseNameUpdated, databaseDescUpdated, deletionProtectionUpdated, labelKeyUpdated, labelValueUpdated, &database),
		},
	})
}

func TestAccYandexYDBDatabaseDedicated_full(t *testing.T) {
	var database ydb.Database
	params := testYandexYDBDatabaseDedicatedParameters{}
	params.name = acctest.RandomWithPrefix("tf-ydb-database-dedicated")
	params.desc = acctest.RandomWithPrefix("tf-ydb-database-dedicated-desc")
	params.labelKey = acctest.RandomWithPrefix("tf-ydb-database-dedicated-label")
	params.labelValue = acctest.RandomWithPrefix("tf-ydb-database-dedicated-label-value")
	params.deletionProtection = "true"

	paramsUpdated := testYandexYDBDatabaseDedicatedParameters{}
	paramsUpdated.name = acctest.RandomWithPrefix("tf-ydb-database-dedicated-updated")
	paramsUpdated.desc = acctest.RandomWithPrefix("tf-ydb-database-dedicated-desc-updated")
	paramsUpdated.labelKey = acctest.RandomWithPrefix("tf-ydb-database-dedicated-label-updated")
	paramsUpdated.labelValue = acctest.RandomWithPrefix("tf-ydb-database-dedicated-label-value-updated")
	paramsUpdated.deletionProtection = "false"

	testConfigFunc := func(params testYandexYDBDatabaseDedicatedParameters) resource.TestStep {
		return resource.TestStep{
			Config: testYandexYDBDatabaseDedicatedFull(params),
			Check: resource.ComposeTestCheckFunc(
				testYandexYDBDatabaseDedicatedExists(ydbDatabaseDedicatedResource, &database),
				resource.TestCheckResourceAttr(ydbDatabaseDedicatedResource, "name", params.name),
				resource.TestCheckResourceAttr(ydbDatabaseDedicatedResource, "description", params.desc),
				resource.TestCheckResourceAttr(ydbDatabaseDedicatedResource, "deletion_protection", params.deletionProtection),
				resource.TestCheckResourceAttrSet(ydbDatabaseDedicatedResource, "folder_id"),
				testYandexYDBDatabaseDedicatedContainsLabel(&database, params.labelKey, params.labelValue),
				testAccCheckCreatedAtAttr(ydbDatabaseDedicatedResource),
			),
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexYDBDatabaseDedicatedDestroy,
		Steps: []resource.TestStep{
			testConfigFunc(params),
			testConfigFunc(paramsUpdated),
		},
	})
}

func basicYandexYDBDatabaseDedicatedTestStep(databaseName, databaseDesc, deletionProtection, labelKey, labelValue string, database *ydb.Database) resource.TestStep {
	return resource.TestStep{
		Config: testYandexYDBDatabaseDedicatedBasic(databaseName, databaseDesc, deletionProtection, labelKey, labelValue),
		Check: resource.ComposeTestCheckFunc(
			testYandexYDBDatabaseDedicatedExists(ydbDatabaseDedicatedResource, database),
			resource.TestCheckResourceAttr(ydbDatabaseDedicatedResource, "name", databaseName),
			resource.TestCheckResourceAttr(ydbDatabaseDedicatedResource, "description", databaseDesc),
			resource.TestCheckResourceAttr(ydbDatabaseDedicatedResource, "deletion_protection", deletionProtection),
			resource.TestCheckResourceAttrSet(ydbDatabaseDedicatedResource, "folder_id"),
			testYandexYDBDatabaseDedicatedContainsLabel(database, labelKey, labelValue),
			testAccCheckCreatedAtAttr(ydbDatabaseDedicatedResource),
		),
	}
}

func testYandexYDBDatabaseDedicatedDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_ydb_database_dedicated" {
			continue
		}

		_, err := testGetYDBDatabaseDedicatedByID(config, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("YDB dedicated database still exists")
		}
	}

	return nil
}

func testYandexYDBDatabaseDedicatedExists(name string, database *ydb.Database) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := testGetYDBDatabaseDedicatedByID(config, rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("YDB dedicated database not found")
		}

		*database = *found
		return nil
	}
}

func testGetYDBDatabaseDedicatedByID(config *Config, ID string) (*ydb.Database, error) {
	req := ydb.GetDatabaseRequest{
		DatabaseId: ID,
	}

	return config.sdk.YDB().Database().Get(context.Background(), &req)
}

func testYandexYDBDatabaseDedicatedContainsLabel(database *ydb.Database, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := database.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

const ydbDatabaseDedicatedDependencies = `
resource "yandex_vpc_network" "ydb-db-dedicated-test-net" {}

resource "yandex_vpc_subnet" "ydb-db-dedicated-test-subnet-a" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.ydb-db-dedicated-test-net.id}"
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "ydb-db-dedicated-test-subnet-b" {
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.ydb-db-dedicated-test-net.id}"
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "ydb-db-dedicated-test-subnet-c" {
  zone           = "ru-central1-c"
  network_id     = "${yandex_vpc_network.ydb-db-dedicated-test-net.id}"
  v4_cidr_blocks = ["10.3.0.0/24"]
}
`

func testYandexYDBDatabaseDedicatedBasic(name string, desc string, deletionProtection, labelKey string, labelValue string) string {
	return fmt.Sprintf(ydbDatabaseDedicatedDependencies+`
resource "yandex_ydb_database_dedicated" "test-ydb-database-dedicated" {
  name        = "%s"
  description = "%s"

  deletion_protection = %s

  labels = {
    %s          = "%s"
    empty-label = ""
  }

  resource_preset_id = "medium"

  scale_policy {
    fixed_scale {
      size = 1
    }
  }

  storage_config {
    group_count     = 1
    storage_type_id = "ssd"
  }

  network_id = "${yandex_vpc_network.ydb-db-dedicated-test-net.id}"
  subnet_ids = [
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-a.id}",
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-b.id}",
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-c.id}",
  ]
}
`, name, desc, deletionProtection, labelKey, labelValue)
}

type testYandexYDBDatabaseDedicatedParameters struct {
	name               string
	desc               string
	labelKey           string
	labelValue         string
	deletionProtection string
}

func testYandexYDBDatabaseDedicatedFull(params testYandexYDBDatabaseDedicatedParameters) string {
	return fmt.Sprintf(ydbDatabaseDedicatedDependencies+`
resource "yandex_ydb_database_dedicated" "test-ydb-database-dedicated" {
  name        = "%s"
  description = "%s"

  deletion_protection = %s

  labels = {
    %s          = "%s"
    empty-label = ""
  }

  resource_preset_id  = "medium"

  scale_policy {
    fixed_scale {
      size = 1
    }
  }

  storage_config {
    group_count     = 1
    storage_type_id = "ssd"
  }

  network_id = "${yandex_vpc_network.ydb-db-dedicated-test-net.id}"
  subnet_ids = [
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-a.id}",
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-b.id}",
    "${yandex_vpc_subnet.ydb-db-dedicated-test-subnet-c.id}",
  ]
}
`,
		params.name,
		params.desc,
		params.deletionProtection,
		params.labelKey,
		params.labelValue)
}
