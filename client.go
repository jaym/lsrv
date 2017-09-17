package lsrv

import (
	"fmt"
	"log"
	"net"
	"strconv"
)

type Client struct {
	manager *ServiceManager
}

func NewClient(state_file string, ip_block *net.IPNet, hosts_file string) *Client {
	client := new(Client)

	client.manager = NewServiceManager(state_file, ip_block, hosts_file)
	return client
}

func (client *Client) Add(service_name string, service_address string, service_port string, dest_port string) {
	service_port_i, err := strconv.ParseUint(service_port, 10, 16)

	if err != nil {
		log.Fatal("Could not parse service port: ", err)
	}

	dest_port_i, err := strconv.ParseUint(dest_port, 10, 16)

	if err != nil {
		log.Fatal("Could not parse expose port: ", err)
	}

	entry, err := client.manager.Add(service_name, service_address, uint16(service_port_i), uint16(dest_port_i))

	if err != nil {
		log.Fatal("Could not add service entry: ", err)
	}

	fmt.Printf("%s.svc %s:%s\n", service_name, entry.DestAddress, dest_port)

}

func (client *Client) Delete(service_name string) {
	err := client.manager.Delete(service_name)
	if err != nil {
		log.Fatalf("Could not delete %s: %s", service_name, err)
	} else {
		fmt.Printf("Removed %s\n", service_name)
	}
}

func (client *Client) Resolve(service_name string) {
	entry, err := client.manager.GetServiceEntry(service_name)
	if err != nil {
		log.Fatalf("Could not resolve %s: %s", service_name, err)
	} else {
		fmt.Printf("%s.svc %s:%d\n", service_name, entry.DestAddress, entry.DestPort)
	}
}

func (client *Client) Restore() {
	services, err := client.manager.Restore()
	if err != nil {
		log.Fatalf("Failed to restore: %s", err)
	} else {
		for service_name, entry := range services {
			fmt.Printf("Restored %s.svc %s:%d\n", service_name, entry.DestAddress, entry.DestPort)
		}
	}
}

func (client *Client) Cleanup() {
	err := client.manager.Cleanup()
	if err != nil {
		log.Fatalf("Failed to cleanup: %s", err)
	}
}
