package wadm

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"go.wasmcloud.dev/x/wasmbus"
	"go.wasmcloud.dev/x/wasmbus/wasmbustest"
)

var testDataPath = path.Join(".", "testdata")

const helloComponent = "ghcr.io/wasmcloud/components/http-hello-world-rust:0.1.0"

func createApp(c *Client, name string) error {
	manifest := newAppManifest(name)
	resp, err := c.ModelPut(context.TODO(), &ModelPutRequest{Manifest: *manifest})
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("error creating app: %v", resp.Message)
	}

	return nil
}

func newAppManifest(name string) *Manifest {
	metadata := ManifestMetadata{
		Name: name,
		Annotations: map[string]string{
			"description": "test app",
		},
	}
	spec := ManifestSpec{
		Components: []Component{
			{
				Name: "hello",
				Type: ComponentTypeComponent,
				Properties: ComponentProperties{
					Image: helloComponent,
				},
			},
		},
	}
	return &Manifest{
		APIVersion: DefaultManifestAPIVersion,
		Kind:       DefaultManifestKind,
		Metadata:   metadata,
		Spec:       spec,
	}
}

func loadTestdata(filePath string) ([]byte, error) {
	fullPath := path.Join(testDataPath, filePath)
	return os.ReadFile(fullPath)
}

func listTestdata(filePath string, pattern string) []string {
	files, err := filepath.Glob(path.Join(testDataPath, filePath, pattern))
	if err != nil {
		panic(err)
	}
	for i, file := range files {
		files[i] = strings.TrimPrefix(file, testDataPath+"/")
	}

	return files
}

func TestClientNats(t *testing.T) {
	nc, teardown := wasmbustest.WithWash(t)
	defer teardown(t)

	bus := wasmbus.NewNatsBus(nc)
	c := NewClient(bus, "default")

	t.Run("ModelList", wrapTest(c, testModelList))
	t.Run("ModelGet", wrapTest(c, testModelGet))
	t.Run("ModelStatus", wrapTest(c, testModelStatus))
	t.Run("ModelPut", wrapTest(c, testModelPut))
	t.Run("ModelVersions", wrapTest(c, testModelVersions))
	t.Run("ModelDelete", wrapTest(c, testModelDelete))
	t.Run("ModelDeploy", wrapTest(c, testModelDeploy))
	t.Run("ModelUndeploy", wrapTest(c, testModelUndeploy))
}

func wrapTest(c *Client, f func(*testing.T, *Client)) func(*testing.T) {
	return func(t *testing.T) {
		f(t, c)
	}
}

func testModelList(t *testing.T, c *Client) {
	if err := createApp(c, "test-list-1"); err != nil {
		t.Fatalf("failed to create app: %v", err)
	}
	if err := createApp(c, "test-list-2"); err != nil {
		t.Fatalf("failed to create app: %v", err)
	}
	if err := createApp(c, "test-list-3"); err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	resp, err := c.ModelList(context.TODO(), &ModelListRequest{})
	if err != nil {
		t.Fatalf("%v", err)
	}

	if want, got := false, resp.IsError(); got != want {
		t.Fatalf("want %v, got %v: %v", want, got, resp.Message)
	}
}

func testModelGet(t *testing.T, c *Client) {
	if err := createApp(c, "test-get"); err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	resp, err := c.ModelGet(context.TODO(), &ModelGetRequest{Name: "test-get", Version: LatestVersion})
	if err != nil {
		t.Fatalf("%v", err)
	}

	if want, got := false, resp.IsError(); got != want {
		t.Fatalf("want %v, got %v: %v", want, got, resp.Message)
	}
}

func testModelStatus(t *testing.T, c *Client) {
	if err := createApp(c, "test-status"); err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	resp, err := c.ModelStatus(context.TODO(), &ModelStatusRequest{Name: "test-status"})
	if err != nil {
		t.Fatalf("%v", err)
	}

	if want, got := false, resp.IsError(); got != want {
		t.Fatalf("want %v, got %v: %v", want, got, resp.Message)
	}
}

func testModelPut(t *testing.T, c *Client) {
	manifest := newAppManifest("test-put")
	resp, err := c.ModelPut(context.TODO(), &ModelPutRequest{Manifest: *manifest})
	if err != nil {
		t.Fatalf("%v", err)
	}

	if want, got := false, resp.IsError(); got != want {
		t.Fatalf("want %v, got %v: %v", want, got, resp.Message)
	}
}

func testModelVersions(t *testing.T, c *Client) {
	if err := createApp(c, "test-versions"); err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	resp, err := c.ModelVersions(context.TODO(), &ModelVersionsRequest{Name: "test-versions"})
	if err != nil {
		t.Fatalf("%v", err)
	}

	if want, got := false, resp.IsError(); got != want {
		t.Fatalf("want %v, got %v: %v", want, got, resp.Message)
	}
}

func testModelDelete(t *testing.T, c *Client) {
	if err := createApp(c, "test-delete"); err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	resp, err := c.ModelDelete(context.TODO(), &ModelDeleteRequest{Name: "test-delete", Version: LatestVersion})
	if err != nil {
		t.Fatalf("%v", err)
	}

	if want, got := false, resp.IsError(); got != want {
		t.Fatalf("want %v, got %v: %v", want, got, resp.Message)
	}
}

func testModelDeploy(t *testing.T, c *Client) {
	if err := createApp(c, "test-deploy"); err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	resp, err := c.ModelDeploy(context.TODO(), &ModelDeployRequest{Name: "test-deploy", Version: LatestVersion})
	if err != nil {
		t.Fatalf("%v", err)
	}

	if want, got := false, resp.IsError(); got != want {
		t.Fatalf("want %v, got %v: %v", want, got, resp.Message)
	}
}

func testModelUndeploy(t *testing.T, c *Client) {
	if err := createApp(c, "test-undeploy"); err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	resp, err := c.ModelUndeploy(context.TODO(), &ModelUndeployRequest{Name: "test-undeploy"})
	if err != nil {
		t.Fatalf("%v", err)
	}

	if want, got := false, resp.IsError(); got != want {
		t.Fatalf("want %v, got %v: %v", want, got, resp.Message)
	}
}
