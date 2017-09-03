package lsrv

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
)

type ServiceManager struct {
	// ip_allocator *IPAllocator
	services   map[string]ServiceEntry
	state_path string
	start_ip   net.IP
	ipt_man    *IPTablesManager
}

type ServiceEntry struct {
	ServiceAddress string
	ServicePort    uint16

	// The service will respond to the address/port below
	DestAddress string
	DestPort    uint16
}

func NewServiceManager(state_path string, start_ip net.IP) *ServiceManager {
	manager := new(ServiceManager)
	manager.state_path = state_path
	manager.start_ip = start_ip

	ipt_man, err := NewIPTablesManager()

	if err != nil {
		log.Fatal(err)
	}

	manager.ipt_man = ipt_man

	if _, err := os.Stat(state_path); os.IsNotExist(err) {
		manager.services = make(map[string]ServiceEntry)
	} else {
		manager.services = load(state_path)
	}

	manager.ipt_man.Cleanup()
	manager.ipt_man.Initialize()

	// TODO: What if start_ip has changed
	for service_name, entry := range manager.services {
		log.Printf("Loaded %s: %+v\n", service_name, entry)
		manager.ipt_man.AddRule(entry.ServiceAddress, entry.ServicePort, entry.DestAddress, entry.DestPort)
	}

	return manager
}

func (manager *ServiceManager) Add(service_name string, service_address string,
	service_port uint16, dest_port uint16) (ServiceEntry, error) {

	entry, present := manager.services[service_name]
	if present {
		return entry, fmt.Errorf("Entry for service %s already exists", service_name)
	}

	entry = ServiceEntry{
		ServiceAddress: service_address,
		ServicePort:    service_port,
		DestAddress:    ipAdd(manager.start_ip, len(manager.services)).String(),
		DestPort:       dest_port,
	}

	manager.services[service_name] = entry
	manager.serialize()
	manager.ipt_man.AddRule(entry.ServiceAddress, entry.ServicePort, entry.DestAddress, entry.DestPort)
	return entry, nil
}

func (manager *ServiceManager) Delete(service_name string) error {
	entry, err := manager.GetServiceEntry(service_name)

	if err != nil {
		return err
	}

	err = manager.ipt_man.RemoveRule(entry.ServiceAddress, entry.ServicePort,
		entry.DestAddress, entry.DestPort)

	if err == nil {
		delete(manager.services, service_name)
		manager.serialize()
	}

	return err
}

func (manager *ServiceManager) List() []ServiceEntry {
	tmp := make([]ServiceEntry, 0, len(manager.services))

	for _, value := range manager.services {
		tmp = append(tmp, value)
	}

	return tmp
}

func (manager *ServiceManager) GetServiceEntry(service_name string) (ServiceEntry, error) {
	entry, present := manager.services[service_name]

	if present {
		return entry, nil
	} else {
		return ServiceEntry{}, fmt.Errorf("Service not found")
	}
}

func (manager *ServiceManager) serialize() {
	services_json, _ := json.Marshal(manager.services)
	// This is not safe. This file should be moved into place
	err := ioutil.WriteFile(manager.state_path, services_json, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func load(path string) map[string]ServiceEntry {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	var services map[string]ServiceEntry
	json.Unmarshal(raw, &services)

	return services
}

func ipAdd(start net.IP, add int) net.IP { // IPv4 only
	start = start.To4()
	result := make(net.IP, 4)
	binary.BigEndian.PutUint32(result, binary.BigEndian.Uint32(start)+uint32(add))
	return result
}
