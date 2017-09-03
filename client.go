package lsrv

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type Client struct {
	socket string
}

func NewClient(socket string) *Client {
	client := new(Client)
	client.socket = socket

	return client
}

func (client *Client) Add(service_name string, service_address string, service_port string, dest_port string) {
	resp, err := client.execute_command(fmt.Sprintf("ADD %s %s %s %s", service_name, service_address,
		service_port, dest_port))

	if err != nil {
		log.Fatal(err)
	} else {
		log.Println(resp)
	}

}

func (client *Client) Delete(service_name string) {
	resp, err := client.execute_command(fmt.Sprintf("DELETE %s", service_name))

	if err != nil {
		log.Fatal(err)
	} else {
		log.Println(resp)
	}

}

func (client *Client) Resolve(service_name string) {
	resp, err := client.execute_command(fmt.Sprintf("GETHOSTBYNAME %s", service_name))

	if err != nil {
		log.Fatal(err)
	} else {
		log.Println(resp)
	}

}

func (client *Client) execute_command(command string) (string, error) {
	conn, err := net.Dial("unix", client.socket)

	if err != nil {
		return "", err
	}

	defer conn.Close()

	fmt.Fprintln(conn, command)

	input := bufio.NewScanner(conn)
	if input.Scan() {
		return input.Text(), nil
	} else {
		return "", fmt.Errorf("Could not read response")
	}
}
