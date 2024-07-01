package main

import (
	"errors"
	"flag"
	"log"
	"slices"
	"strings"

	"github.com/CloudAceTW/go-gce-metadata/service"
	"google.golang.org/api/compute/v1"
)

const (
	ProjectID  = "my-project-id"
	DefultUser = "user"
)

type ChannelObj struct {
	Status bool
	Error  error
}

func main() {
	projectID := flag.String("projectID", ProjectID, "GCP Project ID")
	userSting := flag.String("users", DefultUser, "users list, separated by comma ex: user1,user2")
	flag.Parse()

	removedUsers := strings.Split(*userSting, ",")
	log.Printf("will be clean those users' keys: %+v", removedUsers)

	computeServiceInterface, err := service.NewComputeService()
	if err != nil {
		log.Printf("Failed to create compute service: %v", err)
		return
	}

	log.Printf("clean start")
	DoRemoveSshKey(computeServiceInterface, *projectID, removedUsers)
	log.Printf("clean doen")
}

func DoRemoveSshKey(computeService service.ComputeServiceInterface, projectID string, removedUsers []string) {
	zoneList, err := computeService.GetZones(projectID)
	if err != nil {
		log.Printf("Failed to list zones: %v", err)
		return
	}
	zoneChannel := make(chan ChannelObj, len(zoneList.Items))
	for _, zone := range zoneList.Items {
		go RemovedSshKeyFromZone(computeService, projectID, zone, removedUsers, zoneChannel)
	}
	for i := 0; i < len(zoneList.Items); i++ {
		channelObj := <-zoneChannel
		if !channelObj.Status {
			log.Printf("Failed to remove SSH key: %v", channelObj.Error)
		}
	}
}

func RemovedSshKeyFromZone(computeService service.ComputeServiceInterface, projectID string, zone *compute.Zone, removedUsers []string, c chan ChannelObj) {
	instances, err := computeService.GetInstances(projectID, zone.Name)
	if err != nil {
		log.Printf("Failed to list instances: %v", err)
		c <- ChannelObj{Status: true}
		return
	}
	if len(instances.Items) == 0 {
		log.Printf("No instances found in zone %s", zone.Name)
		c <- ChannelObj{Status: true}
		return
	}
	log.Printf("Start to clean keys in zone %s", zone.Name)
	instanceChannel := make(chan ChannelObj, len(instances.Items))
	for _, instance := range instances.Items {
		log.Printf("Instance: %s", instance.Name)
		go RemovedSshKeyFromInstance(computeService, projectID, zone.Name, instance, removedUsers, instanceChannel)
	}
	var errs []error
	for i := 0; i < len(instances.Items); i++ {
		channelObj := <-instanceChannel
		if !channelObj.Status {
			log.Printf("Failed to remove SSH key: %v", channelObj.Error)
			errs = append(errs, channelObj.Error)
		}
	}
	if len(errs) > 0 {
		c <- ChannelObj{Status: false, Error: errors.New("failed to remove ssh key")}
		return
	}
	c <- ChannelObj{Status: true}
}

func RemovedSshKeyFromInstance(computeService service.ComputeServiceInterface, projectID, zone string, instance *compute.Instance, removedUsers []string, c chan ChannelObj) {
	var newItems []*compute.MetadataItems
	for _, item := range instance.Metadata.Items {
		if item.Key == "ssh-keys" {
			newSshKeyValue := RemoveUserKey(*item.Value, removedUsers)
			item.Value = &newSshKeyValue
		}
		newItems = append(newItems, item)
	}

	metadata := compute.Metadata{
		Fingerprint: instance.Metadata.Fingerprint,
		Items:       newItems,
	}

	err := computeService.SetInstanceMetadata(projectID, zone, instance.Name, &metadata)
	if err != nil {
		log.Printf("Failed to set SSH key: %v", err)
		c <- ChannelObj{Status: false, Error: err}
		return
	}
	c <- ChannelObj{Status: true}
}

func RemoveUserKey(key string, users []string) string {
	var newKeyLists []string
	keyLists := strings.Split(key, "\n")
	for _, key := range keyLists {
		user := strings.Split(key, ":")[0]
		if slices.Contains(users, user) {
			log.Printf("remove key: %s", key)
			continue
		}
		newKeyLists = append(newKeyLists, key)
	}
	newSshKeyValue := strings.Join(newKeyLists, "\n")
	return newSshKeyValue
}
