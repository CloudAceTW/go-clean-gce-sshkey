package service

import (
	"context"

	"google.golang.org/api/compute/v1"
)

type ComputeServiceInterface interface {
	GetService() *compute.Service
	GetZones(projectID string) (*compute.ZoneList, error)
	GetInstances(projectID, zoneName string) (*compute.InstanceList, error)
	SetInstanceMetadata(projectID, zoneName, instanceName string, metadata *compute.Metadata) error
}

type ComputeService struct {
	Service *compute.Service
}

func NewComputeService() (computeServiceInterface ComputeServiceInterface, err error) {
	ctx := context.Background()
	gcpComputeService, err := compute.NewService(ctx)
	if err != nil {
		return computeServiceInterface, err
	}
	computeService := ComputeService{
		Service: gcpComputeService,
	}
	computeServiceInterface = &computeService
	return computeServiceInterface, nil
}

func (cs *ComputeService) GetZones(projectID string) (*compute.ZoneList, error) {
	zoneList, err := cs.Service.Zones.List(projectID).Do()
	return zoneList, err
}

func (cs *ComputeService) GetInstances(projectID, zoneName string) (*compute.InstanceList, error) {
	instanceList, err := cs.Service.Instances.List(projectID, zoneName).Do()
	return instanceList, err
}

func (cs *ComputeService) SetInstanceMetadata(projectID, zoneName, instanceName string, metadata *compute.Metadata) error {
	_, err := cs.Service.Instances.SetMetadata(projectID, zoneName, instanceName, metadata).Do()
	return err
}

func (cs *ComputeService) GetService() *compute.Service {
	return cs.Service
}
