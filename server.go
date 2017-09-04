package lsrv

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

func Spawn(socket_path string) error {
	ln, err := net.Listen("unix", socket_path)

	if err != nil {
		log.Fatal("Listen error: ", err)
		return err
	}

	start_ip := net.IPv4(172, 22, 0, 1)

	manager := NewServiceManager("./state", start_ip)

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
				fmt.Fprintf(fd, "ERROR Could not parse command")
			} else {
				service_port, serr := strconv.ParseUint(command_split[3], 10, 16)
				dest_port, derr := strconv.ParseUint(command_split[4], 10, 16)

				if serr != nil || derr != nil {
					log.Println("Could not parse port")
					fd.Close()
					break
				}

				entry, err := manager.Add(command_split[1], command_split[2], uint16(service_port), uint16(dest_port))

				if err != nil {
					fmt.Fprintf(fd, "ERROR %s\n", err.Error())
				} else {
					fmt.Fprintf(fd, "OK %s\n", entry.DestAddress)
				}

				log.Println("Added ", entry)
			}
		case "DELETE":
			if len(command_split) != 2 {
				fmt.Fprintln(fd, "ERROR Could not parse command")
			} else {
				err := manager.Delete(command_split[1])
				if err != nil {
					fmt.Fprintf(fd, "ERROR %s\n", err.Error())
				} else {
					fmt.Fprintln(fd, "OK")
				}
			}
		case "GETHOSTBYNAME":
			if len(command_split) != 2 {
				fmt.Fprintln(fd, "ERROR Could not parse command")
			} else {
				entry, err := manager.GetServiceEntry(command_split[1])
				if err != nil {
					fmt.Fprintf(fd, "ERROR %s\n", err.Error())
				} else {
					fmt.Fprintf(fd, "OK %s\n", entry.DestAddress)
				}
			}
		default:
			log.Println("Unknown command: ", command_split[0])
			fmt.Fprintln(fd, "ERROR Unknown Command")
		}
		fd.Close()
	}

	return nil
}
