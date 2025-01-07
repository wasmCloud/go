package wadm

import (
	"context"
)

type API interface {
	// ModelList returns a list of models
	// wadm.api.{lattice-id}.model.get
	ModelList(ctx context.Context, req *ModelListRequest) (*ModelListResponse, error)
	// ModelGet returns a model by name and version
	// wadm.api.{lattice-id}.model.get.{name}
	ModelGet(ctx context.Context, req *ModelGetRequest) (*ModelGetResponse, error)
	// ModelVersions returns a list of versions for a model
	// wadm.api.{lattice-id}.model.versions.{name}
	ModelVersions(ctx context.Context, req *ModelVersionsRequest) (*ModelVersionsResponse, error)
	// ModelStatus returns the status of a model
	// wadm.api.{lattice-id}.model.status.{name}
	ModelStatus(ctx context.Context, req *ModelStatusRequest) (*ModelStatusResponse, error)
	// ModelPut creates or updates a model
	// wadm.api.{lattice-id}.model.put
	ModelPut(ctx context.Context, req *ModelPutRequest) (*ModelPutResponse, error)
	// ModelDelete deletes a model
	// wadm.api.{lattice-id}.model.del.{name}
	ModelDelete(ctx context.Context, req *ModelDeleteRequest) (*ModelDeleteResponse, error)
	// ModelDeploy deploys a model
	// wadm.api.{lattice-id}.model.deploy.{name}
	ModelDeploy(ctx context.Context, req *ModelDeployRequest) (*ModelDeployResponse, error)
	// ModelUndeploy undeploys a model
	// wadm.api.{lattice-id}.model.undeploy.{name}
	ModelUndeploy(ctx context.Context, req *ModelUndeployRequest) (*ModelUndeployResponse, error)
}

type APIMock struct {
	ModelListFunc     func(ctx context.Context, req *ModelListRequest) (*ModelListResponse, error)
	ModelGetFunc      func(ctx context.Context, req *ModelGetRequest) (*ModelGetResponse, error)
	ModelStatusFunc   func(ctx context.Context, req *ModelStatusRequest) (*ModelStatusResponse, error)
	ModelVersionsFunc func(ctx context.Context, req *ModelVersionsRequest) (*ModelVersionsResponse, error)
	ModelPutFunc      func(ctx context.Context, req *ModelPutRequest) (*ModelPutResponse, error)
	ModelDeleteFunc   func(ctx context.Context, req *ModelDeleteRequest) (*ModelDeleteResponse, error)
	ModelDeployFunc   func(ctx context.Context, req *ModelDeployRequest) (*ModelDeployResponse, error)
	ModelUndeployFunc func(ctx context.Context, req *ModelUndeployRequest) (*ModelUndeployResponse, error)
}

var _ API = (*APIMock)(nil)

func (m *APIMock) ModelList(ctx context.Context, req *ModelListRequest) (*ModelListResponse, error) {
	return m.ModelListFunc(ctx, req)
}

func (m *APIMock) ModelGet(ctx context.Context, req *ModelGetRequest) (*ModelGetResponse, error) {
	return m.ModelGetFunc(ctx, req)
}

func (m *APIMock) ModelStatus(ctx context.Context, req *ModelStatusRequest) (*ModelStatusResponse, error) {
	return m.ModelStatusFunc(ctx, req)
}

func (m *APIMock) ModelVersions(ctx context.Context, req *ModelVersionsRequest) (*ModelVersionsResponse, error) {
	return m.ModelVersionsFunc(ctx, req)
}

func (m *APIMock) ModelPut(ctx context.Context, req *ModelPutRequest) (*ModelPutResponse, error) {
	return m.ModelPutFunc(ctx, req)
}

func (m *APIMock) ModelDelete(ctx context.Context, req *ModelDeleteRequest) (*ModelDeleteResponse, error) {
	return m.ModelDeleteFunc(ctx, req)
}

func (m *APIMock) ModelDeploy(ctx context.Context, req *ModelDeployRequest) (*ModelDeployResponse, error) {
	return m.ModelDeployFunc(ctx, req)
}

func (m *APIMock) ModelUndeploy(ctx context.Context, req *ModelUndeployRequest) (*ModelUndeployResponse, error) {
	return m.ModelUndeployFunc(ctx, req)
}

const (
	VersionAnnotation = "version"
)
