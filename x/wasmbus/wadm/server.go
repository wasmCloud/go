package wadm

import (
	"context"
	"strings"

	"go.wasmcloud.dev/x/wasmbus"
)

type Server struct {
	*wasmbus.Server
	Lattice string
	api     API
}

func NewServer(bus wasmbus.Bus, lattice string, api API) *Server {
	return &Server{
		Server:  wasmbus.NewServer(bus),
		Lattice: lattice,
		api:     api,
	}
}

func (s *Server) Serve() error {
	listModelLegacy := wasmbus.NewRequestHandler(ModelListRequest{}, ModelListResponse{}, s.api.ModelList)
	listModelLegacy.PostRequest = func(_ context.Context, resp *ModelListResponse, msg *wasmbus.Message) error {
		for i := range resp.Models {
			resp.Models[i].Status = resp.Models[i].DetailedStatus.Info.Type
		}
		tmpMsg, err := wasmbus.Encode(msg.Subject, resp.Models)
		if err != nil {
			return err
		}
		msg.Data = tmpMsg.Data

		return nil
	}

	if err := s.RegisterHandler(s.subject("model", "list"), listModelLegacy); err != nil {
		return err
	}

	listModel := wasmbus.NewRequestHandler(ModelListRequest{}, ModelListResponse{}, s.api.ModelList)
	if err := s.RegisterHandler(s.subject("model", "get"), listModel); err != nil {
		return err
	}

	putModel := wasmbus.NewRequestHandler(ModelPutRequest{}, ModelPutResponse{}, s.api.ModelPut)
	if err := s.RegisterHandler(s.subject("model", "put"), putModel); err != nil {
		return err
	}

	getModel := wasmbus.NewRequestHandler(ModelGetRequest{}, ModelGetResponse{}, s.api.ModelGet)
	getModel.PreRequest = func(_ context.Context, req *ModelGetRequest, msg *wasmbus.Message) error {
		req.Name = msg.LastSubjectPart()
		return nil
	}
	if err := s.RegisterHandler(s.subject("model", "get", wasmbus.PatternAll), getModel); err != nil {
		return err
	}

	statusModel := wasmbus.NewRequestHandler(ModelStatusRequest{}, ModelStatusResponse{}, s.api.ModelStatus)
	statusModel.PreRequest = func(_ context.Context, req *ModelStatusRequest, msg *wasmbus.Message) error {
		req.Name = msg.LastSubjectPart()
		return nil
	}
	if err := s.RegisterHandler(s.subject("model", "status", wasmbus.PatternAll), statusModel); err != nil {
		return err
	}

	delModel := wasmbus.NewRequestHandler(ModelDeleteRequest{}, ModelDeleteResponse{}, s.api.ModelDelete)
	delModel.PreRequest = func(_ context.Context, req *ModelDeleteRequest, msg *wasmbus.Message) error {
		req.Name = msg.LastSubjectPart()
		return nil
	}
	if err := s.RegisterHandler(s.subject("model", "del", wasmbus.PatternAll), delModel); err != nil {
		return err
	}

	versionsModel := wasmbus.NewRequestHandler(ModelVersionsRequest{}, ModelVersionsResponse{}, s.api.ModelVersions)
	versionsModel.PreRequest = func(_ context.Context, req *ModelVersionsRequest, msg *wasmbus.Message) error {
		req.Name = msg.LastSubjectPart()
		return nil
	}
	if err := s.RegisterHandler(s.subject("model", "versions", wasmbus.PatternAll), versionsModel); err != nil {
		return err
	}

	deployModel := wasmbus.NewRequestHandler(ModelDeployRequest{}, ModelDeployResponse{}, s.api.ModelDeploy)
	deployModel.PreRequest = func(_ context.Context, req *ModelDeployRequest, msg *wasmbus.Message) error {
		req.Name = msg.LastSubjectPart()
		return nil
	}
	if err := s.RegisterHandler(s.subject("model", "deploy", wasmbus.PatternAll), deployModel); err != nil {
		return err
	}

	undeployModel := wasmbus.NewRequestHandler(ModelUndeployRequest{}, ModelUndeployResponse{}, s.api.ModelUndeploy)
	undeployModel.PreRequest = func(_ context.Context, req *ModelUndeployRequest, msg *wasmbus.Message) error {
		req.Name = msg.LastSubjectPart()
		return nil
	}
	if err := s.RegisterHandler(s.subject("model", "undeploy", wasmbus.PatternAll), undeployModel); err != nil {
		return err
	}

	return nil
}

func (s *Server) subject(ids ...string) string {
	parts := append([]string{wasmbus.PrefixWadm, s.Lattice}, ids...)
	return strings.Join(parts, ".")
}
