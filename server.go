package lsrv

import (
	"bufio"
	"encoding/binary"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/coreos/go-iptables/iptables"
)

type ServiceEntry struct {
	service_name    string
	service_address string
	service_port    uint16

	// The service will respond to the address/port below
	dest_address string
	dest_port    uint16
}

func Spawn(socket_path string) error {
	ipt, iperr := initializeIpTables()

	if iperr != nil {
		log.Fatal("Could not initialize IPTables: ", iperr)
	}

	ln, err := net.Listen("unix", socket_path)

	if err != nil {
		log.Fatal("Listen error: ", err)
		return err
	}

	services := []ServiceEntry{}
	start_ip := net.IPv4(172, 22, 0, 1)

	for {
		//TODO: err check
		fd, _ := ln.Accept()
		input := bufio.NewScanner(fd)

		success := input.Scan()
		if !success {
			log.Println(input.Err())
			fd.Close()
		}

		command := input.Text()
		command_split := strings.Fields(command)

		if len(command_split) == 0 {
			log.Println("Could not parse command")
			fd.Close()
		}

		switch command_split[0] {
		case "ADD":
			// ADD service_name service_ip service_port dest_port
			// Returns:
			//   OK
			if len(command_split) != 5 {
				log.Println("Could not parse command")
				fd.Close()
			} else {
				service_port, serr := strconv.ParseUint(command_split[3], 10, 16)
				dest_port, derr := strconv.ParseUint(command_split[4], 10, 16)

				if serr != nil || derr != nil {
					log.Println("Could not parse port")
					fd.Close()
					break
				}

				entry := ServiceEntry{
					service_name:    command_split[1],
					service_address: command_split[2],
					service_port:    uint16(service_port),
					dest_address:    ipAdd(start_ip, len(services)).String(),
					dest_port:       uint16(dest_port),
				}

				services = append(services, entry)
				err = materialize(ipt, &services)

				if err != nil {
					log.Println(err)
					break
				}

				log.Println("Added ", entry)
			}
		default:
			log.Println("Unknown command: ", command_split[0])
		}
		fd.Close()
	}

	return nil
}

func materialize(ipt *iptables.IPTables, services *[]ServiceEntry) error {
	// Recreate the chain. This is wasteful and not zero downtime, but
	// it's easy
	err := createChain(ipt)
	if err != nil {
		return err
	}

	for _, entry := range *services {
		rulespec := []string{"-p", "tcp", "-d", entry.dest_address, "--dport",
			strconv.FormatUint(uint64(entry.dest_port), 10), "-j", "DNAT",
			"--to", entry.service_address + ":" + strconv.FormatUint(uint64(entry.service_port), 10)}
		err = ipt.Append("nat", "LSRV", rulespec...)
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func ipAdd(start net.IP, add int) net.IP { // IPv4 only
	start = start.To4()
	result := make(net.IP, 4)
	binary.BigEndian.PutUint32(result, binary.BigEndian.Uint32(start)+uint32(add))
	return result
}

func initializeIpTables() (*iptables.IPTables, error) {
	ipt, err := iptables.New()

	if err != nil {
		return nil, err
	}

	err = createChain(ipt)

	if err != nil {
		return nil, err
	}

	return ipt, nil
}

func createChain(ipt *iptables.IPTables) error {
	cleanup(ipt)

	ipt.NewChain("nat", "LSRV")
	ipt.AppendUnique("nat", "OUTPUT", "-jLSRV")

	return nil
}

func cleanup(ipt *iptables.IPTables) error {
	chains, err := ipt.ListChains("nat")
	if err != nil {
		return err
	}

	containsChain := false
	for _, elem := range chains {
		if elem == "LSRV" {
			containsChain = true
		}
	}

	if containsChain {
		log.Println("Deleting chain LSRV")
		ipt.Delete("nat", "OUTPUT", "-jLSRV")
		ipt.ClearChain("nat", "LSRV")
		ipt.DeleteChain("nat", "LSRV")
	}
	return nil
}
