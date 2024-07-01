package main

import (
	"errors"
	"testing"

	"google.golang.org/api/compute/v1"
)

// MockComputeService is a mock implementation of the ComputeServiceInterface
type MockComputeService struct {
	GetZonesFunc            func(projectID string) (*compute.ZoneList, error)
	GetInstancesFunc        func(projectID, zone string) (*compute.InstanceList, error)
	SetInstanceMetadataFunc func(projectID, zone, instanceName string, metadata *compute.Metadata) error
}

// GetService implements service.ComputeServiceInterface.
func (m *MockComputeService) GetService() *compute.Service {
	panic("unimplemented")
}

// GetZones implements the ComputeServiceInterface interface
func (m *MockComputeService) GetZones(projectID string) (*compute.ZoneList, error) {
	return m.GetZonesFunc(projectID)
}

// GetInstances implements the ComputeServiceInterface interface
func (m *MockComputeService) GetInstances(projectID, zone string) (*compute.InstanceList, error) {
	return m.GetInstancesFunc(projectID, zone)
}

// SetInstanceMetadata implements the ComputeServiceInterface interface
func (m *MockComputeService) SetInstanceMetadata(projectID, zone, instanceName string, metadata *compute.Metadata) error {
	return m.SetInstanceMetadataFunc(projectID, zone, instanceName, metadata)
}

func TestDoRemoveSshKey(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockComputeService := &MockComputeService{
			GetZonesFunc: func(projectID string) (*compute.ZoneList, error) {
				return &compute.ZoneList{
					Items: []*compute.Zone{
						{Name: "zone-1"},
						{Name: "zone-2"},
					},
				}, nil
			},
			GetInstancesFunc: func(projectID, zone string) (*compute.InstanceList, error) {
				return &compute.InstanceList{
					Items: []*compute.Instance{
						{Name: "instance-1", Metadata: &compute.Metadata{Items: []*compute.MetadataItems{{Key: "ssh-keys", Value: ptr("user1:key1\nuser2:key2")}}}},
						{Name: "instance-2", Metadata: &compute.Metadata{Items: []*compute.MetadataItems{{Key: "ssh-keys", Value: ptr("user3:key3\nuser4:key4")}}}},
					},
				}, nil
			},
			SetInstanceMetadataFunc: func(projectID, zone, instanceName string, metadata *compute.Metadata) error {
				return nil
			},
		}
		DoRemoveSshKey(mockComputeService, ProjectID, []string{"user1", "user3"})
	})

	t.Run("ErrorGettingZones", func(t *testing.T) {
		mockComputeService := &MockComputeService{
			GetZonesFunc: func(projectID string) (*compute.ZoneList, error) {
				return nil, errors.New("error getting zones")
			},
		}
		DoRemoveSshKey(mockComputeService, ProjectID, []string{"user1"})
	})

	t.Run("ErrorGettingInstances", func(t *testing.T) {
		mockComputeService := &MockComputeService{
			GetZonesFunc: func(projectID string) (*compute.ZoneList, error) {
				return &compute.ZoneList{
					Items: []*compute.Zone{
						{Name: "zone-1"},
					},
				}, nil
			},
			GetInstancesFunc: func(projectID, zone string) (*compute.InstanceList, error) {
				return nil, errors.New("error getting instances")
			},
		}
		DoRemoveSshKey(mockComputeService, ProjectID, []string{"user1"})
	})

	t.Run("ErrorSettingInstanceMetadata", func(t *testing.T) {
		mockComputeService := &MockComputeService{
			GetZonesFunc: func(projectID string) (*compute.ZoneList, error) {
				return &compute.ZoneList{
					Items: []*compute.Zone{
						{Name: "zone-1"},
					},
				}, nil
			},
			GetInstancesFunc: func(projectID, zone string) (*compute.InstanceList, error) {
				return &compute.InstanceList{
					Items: []*compute.Instance{
						{Name: "instance-1", Metadata: &compute.Metadata{Items: []*compute.MetadataItems{{Key: "ssh-keys", Value: ptr("user1:key1\nuser2:key2")}}}},
					},
				}, nil
			},
			SetInstanceMetadataFunc: func(projectID, zone, instanceName string, metadata *compute.Metadata) error {
				return errors.New("error setting instance metadata")
			},
		}
		DoRemoveSshKey(mockComputeService, ProjectID, []string{"user1"})
	})
}

func TestRemoveUserKey(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		key := "user1:key1\nuser2:key2\nuser3:key3"
		users := []string{"user1", "user3"}
		newKey := RemoveUserKey(key, users)
		expectedKey := "user2:key2"
		if newKey != expectedKey {
			t.Errorf("Expected key %s, got %s", expectedKey, newKey)
		}
	})

	t.Run("EmptyKey", func(t *testing.T) {
		key := ""
		users := []string{"user1", "user3"}
		newKey := RemoveUserKey(key, users)
		expectedKey := ""
		if newKey != expectedKey {
			t.Errorf("Expected key %s, got %s", expectedKey, newKey)
		}
	})

	t.Run("NoMatchingUsers", func(t *testing.T) {
		key := "user1:key1\nuser2:key2\nuser3:key3"
		users := []string{"user4", "user5"}
		newKey := RemoveUserKey(key, users)
		expectedKey := "user1:key1\nuser2:key2\nuser3:key3"
		if newKey != expectedKey {
			t.Errorf("Expected key %s, got %s", expectedKey, newKey)
		}
	})
}

func TestRemovedSshKeyFromInstance(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockComputeService := &MockComputeService{
			GetZonesFunc: func(projectID string) (*compute.ZoneList, error) {
				return &compute.ZoneList{
					Items: []*compute.Zone{
						{Name: "test-zone"},
					},
				}, nil
			},
			GetInstancesFunc: func(projectID, zone string) (*compute.InstanceList, error) {
				return &compute.InstanceList{
					Items: []*compute.Instance{
						{Name: "test-instance", Metadata: &compute.Metadata{Items: []*compute.MetadataItems{{Key: "ssh-keys", Value: ptr("user1:key1\nuser2:key2")}}}},
					},
				}, nil
			},
			SetInstanceMetadataFunc: func(projectID, zone, instanceName string, metadata *compute.Metadata) error {
				return nil
			},
		}
		instance := &compute.Instance{
			Name: "test-instance",
			Metadata: &compute.Metadata{
				Items: []*compute.MetadataItems{
					{
						Key:   "ssh-keys",
						Value: ptr("user1:key1\nuser2:key2\nuser3:key3"),
					},
				},
			},
		}
		removedUsers := []string{"user1"}
		c := make(chan ChannelObj, 1)
		RemovedSshKeyFromInstance(mockComputeService, "test-project", "test-zone", instance, removedUsers, c)
		channelObj := <-c
		if !channelObj.Status {
			t.Errorf("Expected successful removal, got error: %v", channelObj.Error)
		}
		if instance.Metadata.Items[0].Value == nil {
			t.Error("Expected ssh-key value to be updated, but it's nil")
		} else {
			expectedKey := "user2:key2\nuser3:key3"
			if *instance.Metadata.Items[0].Value != expectedKey {
				t.Errorf("Expected updated ssh-key value to be %s, got %s", expectedKey, *instance.Metadata.Items[0].Value)
			}
		}
	})

	t.Run("FailedToSetInstanceMetadata", func(t *testing.T) {
		mockComputeService := &MockComputeService{
			GetZonesFunc: func(projectID string) (*compute.ZoneList, error) {
				return &compute.ZoneList{
					Items: []*compute.Zone{
						{Name: "test-zone"},
					},
				}, nil
			},
			GetInstancesFunc: func(projectID, zone string) (*compute.InstanceList, error) {
				return &compute.InstanceList{
					Items: []*compute.Instance{
						{Name: "test-instance", Metadata: &compute.Metadata{Items: []*compute.MetadataItems{{Key: "ssh-keys", Value: ptr("user1:key1\nuser2:key2")}}}},
					},
				}, nil
			},
			SetInstanceMetadataFunc: func(projectID, zone, instanceName string, metadata *compute.Metadata) error {
				return errors.New("failed to set instance metadata")
			},
		}
		instance := &compute.Instance{
			Name: "test-instance",
			Metadata: &compute.Metadata{
				Items: []*compute.MetadataItems{
					{
						Key:   "ssh-keys",
						Value: ptr("user1:key1\nuser2:key2"),
					},
				},
			},
		}
		removedUsers := []string{"user1"}
		c := make(chan ChannelObj, 1)
		RemovedSshKeyFromInstance(mockComputeService, "test-project", "test-zone", instance, removedUsers, c)
		channelObj := <-c
		if channelObj.Status {
			t.Errorf("Expected failed removal, got success")
		}
		if channelObj.Error == nil {
			t.Error("Expected error to be returned, but it's nil")
		} else if channelObj.Error.Error() != "failed to set instance metadata" {
			t.Errorf("Expected error message to be 'failed to set instance metadata', got %s", channelObj.Error.Error())
		}
	})
}

func TestRemovedSshKeyFromZone(t *testing.T) {
	t.Run("Success but no instance found", func(t *testing.T) {
		mockService := &MockComputeService{
			GetZonesFunc: func(projectID string) (*compute.ZoneList, error) {
				return &compute.ZoneList{
					Items: []*compute.Zone{
						{Name: "zone1"},
					},
				}, nil
			},
			GetInstancesFunc: func(projectID, zone string) (*compute.InstanceList, error) {
				return &compute.InstanceList{
					Items: []*compute.Instance{},
				}, nil
			},
			SetInstanceMetadataFunc: func(projectID, zone, instanceName string, metadata *compute.Metadata) error {
				return nil
			},
		}

		removedUsers := []string{"user1"}

		zoneChannel := make(chan ChannelObj, 1)
		RemovedSshKeyFromZone(mockService, "my-project-id", &compute.Zone{Name: "zone1"}, removedUsers, zoneChannel)

		channelObj := <-zoneChannel
		if !channelObj.Status {
			t.Errorf("Expected success but got error: %v", channelObj.Error)
		}
	})
	t.Run("Success", func(t *testing.T) {
		mockService := &MockComputeService{
			GetZonesFunc: func(projectID string) (*compute.ZoneList, error) {
				return &compute.ZoneList{
					Items: []*compute.Zone{
						{Name: "zone1"},
						{Name: "zone2"},
					},
				}, nil
			},
			GetInstancesFunc: func(projectID, zone string) (*compute.InstanceList, error) {
				return &compute.InstanceList{
					Items: []*compute.Instance{
						{
							Name: "instance1",
							Metadata: &compute.Metadata{
								Items: []*compute.MetadataItems{
									{
										Key:   "ssh-keys",
										Value: ptr("user1:key1\nuser2:key2\nuser3:key3"),
									},
								},
							},
						},
					},
				}, nil
			},
			SetInstanceMetadataFunc: func(projectID, zone, instanceName string, metadata *compute.Metadata) error {
				return nil
			},
		}

		removedUsers := []string{"user1"}

		zoneChannel := make(chan ChannelObj, 2)
		RemovedSshKeyFromZone(mockService, "my-project-id", &compute.Zone{Name: "zone1"}, removedUsers, zoneChannel)
		RemovedSshKeyFromZone(mockService, "my-project-id", &compute.Zone{Name: "zone2"}, removedUsers, zoneChannel)

		for i := 0; i < 2; i++ {
			channelObj := <-zoneChannel
			if !channelObj.Status {
				t.Errorf("Expected success but got error: %v", channelObj.Error)
			}
		}
	})

	t.Run("FailedToGetInstances", func(t *testing.T) {
		mockService := &MockComputeService{
			GetZonesFunc: func(projectID string) (*compute.ZoneList, error) {
				return &compute.ZoneList{
					Items: []*compute.Zone{
						{Name: "zone1"},
					},
				}, nil
			},
			GetInstancesFunc: func(projectID, zone string) (*compute.InstanceList, error) {
				return nil, errors.New("failed to get instances")
			},
		}

		removedUsers := []string{"user1"}

		zoneChannel := make(chan ChannelObj, 1)
		RemovedSshKeyFromZone(mockService, "my-project-id", &compute.Zone{Name: "zone1"}, removedUsers, zoneChannel)

		channelObj := <-zoneChannel
		if !channelObj.Status {
			t.Logf("Expected error: %v", channelObj.Error)
			return
		}
		t.Errorf("Expected error but got success")
	})

	t.Run("FailedToSetInstanceMetadata", func(t *testing.T) {
		mockService := &MockComputeService{
			GetZonesFunc: func(projectID string) (*compute.ZoneList, error) {
				return &compute.ZoneList{
					Items: []*compute.Zone{
						{Name: "zone1"},
					},
				}, nil
			},
			GetInstancesFunc: func(projectID, zone string) (*compute.InstanceList, error) {
				return &compute.InstanceList{
					Items: []*compute.Instance{
						{
							Name: "instance1",
							Metadata: &compute.Metadata{
								Items: []*compute.MetadataItems{
									{
										Key:   "ssh-keys",
										Value: ptr("user1:key1\nuser2:key2\nuser3:key3"),
									},
								},
							},
						},
					},
				}, nil
			},
			SetInstanceMetadataFunc: func(projectID, zone, instanceName string, metadata *compute.Metadata) error {
				return errors.New("failed to set metadata")
			},
		}

		removedUsers := []string{"user1"}

		zoneChannel := make(chan ChannelObj, 1)
		RemovedSshKeyFromZone(mockService, "my-project-id", &compute.Zone{Name: "zone1"}, removedUsers, zoneChannel)

		channelObj := <-zoneChannel
		if channelObj.Status {
			t.Errorf("Expected error but got success")
		}
	})
}

func ptr(s string) *string {
	return &s
}
