package lsrv

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
)

type ServiceManager struct {
	services   map[string]ServiceEntry
	state_path string
	start_ip   net.IP
	next_ip    string
	free_ips   []string
	ipt_man    *IPTablesManager
}

type ServiceEntry struct {
	ServiceAddress string
	ServicePort    uint16

	// The service will respond to the address/port below
	DestAddress string
	DestPort    uint16
}

type StateFile struct {
	Services map[string]ServiceEntry `json:"services"`
	NextIp   string
	FreeIps  []string
}

func NewServiceManager(state_path string, start_ip net.IP) *ServiceManager {
	manager := new(ServiceManager)
	manager.state_path = state_path
	manager.start_ip = start_ip
	manager.next_ip = start_ip.String()
	manager.services = make(map[string]ServiceEntry)
	manager.free_ips = []string{}

	ipt_man, err := NewIPTablesManager()

	if err != nil {
		log.Fatal(err)
	}

	manager.ipt_man = ipt_man

	if _, err := os.Stat(state_path); !os.IsNotExist(err) {
		state_file := load(state_path)

		if state_file.Services != nil {
			manager.services = state_file.Services
		}

		if state_file.FreeIps != nil {
			manager.free_ips = state_file.FreeIps
		}

		if state_file.NextIp != "" {
			manager.next_ip = state_file.NextIp
		}
	}

	return manager
}

func (manager *ServiceManager) Add(service_name string, service_address string,
	service_port uint16, dest_port uint16) (ServiceEntry, error) {

	entry, present := manager.services[service_name]
	if present {
		return entry, fmt.Errorf("Entry for service %s already exists", service_name)
	}

	var next_ip string

	if len(manager.free_ips) > 0 {
		next_ip, manager.free_ips = manager.free_ips[len(manager.free_ips)-1],
			manager.free_ips[:len(manager.free_ips)-1]
	} else {
		next_ip = manager.next_ip
		log.Printf("next_ip %+v\n", next_ip)
		manager.next_ip = ipAdd(net.ParseIP(next_ip), 1).String()
	}

	entry = ServiceEntry{
		ServiceAddress: service_address,
		ServicePort:    service_port,
		DestAddress:    next_ip,
		DestPort:       dest_port,
	}

	manager.services[service_name] = entry
	manager.serialize()
	manager.ipt_man.AddRule(entry.ServiceAddress, entry.ServicePort, entry.DestAddress, entry.DestPort)
	err := manager.write_etc_hosts(true)

	if err != nil {
		return ServiceEntry{}, nil
	}
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
		manager.free_ips = append(manager.free_ips, entry.DestAddress)
		manager.serialize()
		err = manager.write_etc_hosts(true)
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

func (manager *ServiceManager) Restore() (map[string]ServiceEntry, error) {
	//TODO: Return errors
	manager.ipt_man.Cleanup()
	manager.ipt_man.Initialize()

	// TODO: What if start_ip has changed
	for _, entry := range manager.services {
		manager.ipt_man.AddRule(entry.ServiceAddress, entry.ServicePort, entry.DestAddress, entry.DestPort)
	}

	if err := manager.write_etc_hosts(true); err != nil {
		return nil, err
	}

	return manager.services, nil
}

func (manager *ServiceManager) serialize() {
	services_json, err := json.Marshal(&StateFile{
		Services: manager.services,
		NextIp:   manager.next_ip,
		FreeIps:  manager.free_ips,
	})

	if err != nil {
		log.Fatal(err)
	}

	// This is not safe. This file should be moved into place
	err = ioutil.WriteFile(manager.state_path, services_json, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

// This function is temporary and will be replaced
// by an nsswitch module
func (manager *ServiceManager) write_etc_hosts(include_lsrv bool) error {
	infile, err := os.Open("/etc/hosts")
	defer infile.Close()
	outfile, err := os.Create("/etc/_hosts.lsrv")
	defer outfile.Close()

	if err != nil {
		return err
	}

	input := bufio.NewScanner(infile)
	writer := bufio.NewWriter(outfile)

	for input.Scan() {
		line := input.Text()
		if !strings.Contains(line, "# __lsrv_managed") {
			writer.WriteString(input.Text())
			writer.WriteString("\n")
		}
	}

	if err = input.Err(); err != nil {
		return err
	}

	if include_lsrv {
		for service_name, entry := range manager.services {
			fmt.Fprintf(writer, "%s %s.svc # __lsrv_managed\n", entry.DestAddress, service_name)
		}
	}

	writer.Flush()

	outfile.Close()

	err = os.Rename("/etc/_hosts.lsrv", "/etc/hosts")

	if err != nil {
		return err
	}
	return nil
}

func load(path string) StateFile {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	var state_file StateFile
	json.Unmarshal(raw, &state_file)

	return state_file
}

func ipAdd(start net.IP, add int) net.IP { // IPv4 only
	start = start.To4()
	result := make(net.IP, 4)
	binary.BigEndian.PutUint32(result, binary.BigEndian.Uint32(start)+uint32(add))
	return result
}
