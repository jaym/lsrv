package main

import (
	"os"

	"github.com/jaym/lsrv"
	"github.com/urfave/cli"
)

var socket string

func main() {
	app := cli.NewApp()
	app.Name = "lsrv"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "socket",
			Value:       "/tmp/lsrv.sock",
			Usage:       "UNIX domain socket to use",
			Destination: &socket,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:        "add",
			Usage:       "Add a service to be managed",
			ArgsUsage:   "service_name service_port expose_port",
			Description: "service_name will be assigned an ip address. Any traffic going to service_name:expose_port will be forwarded to 127.0.0.1:service_port",
			Action: func(c *cli.Context) error {
				if len(c.Args()) != 3 {
					cli.ShowCommandHelpAndExit(c, "add", 1)
				}
				args := c.Args()
				client().Add(args[0], "127.0.0.1", args[1], args[2])
				return nil
			},
		},
		{
			Name:        "rm",
			Usage:       "Remove a service that is managed",
			ArgsUsage:   "service_name",
			Description: "service_name will no longer the forwarded",
			Action: func(c *cli.Context) error {
				if len(c.Args()) != 1 {
					cli.ShowCommandHelpAndExit(c, "rm", 1)
				}
				args := c.Args()
				client().Delete(args[0])
				return nil
			},
		},
		{
			Name:        "restore",
			Usage:       "Restore all services",
			Description: "Populates IPTables and the hosts file based on the current state",
			Action: func(c *cli.Context) error {
				if len(c.Args()) != 0 {
					cli.ShowCommandHelpAndExit(c, "restore", 1)
				}
				client().Restore()
				return nil
			},
		},
		{
			Name:      "resolve",
			Usage:     "Resolve the ip address of a service that is managed",
			ArgsUsage: "service_name",
			Action: func(c *cli.Context) error {
				if len(c.Args()) != 1 {
					cli.ShowCommandHelpAndExit(c, "resolve", 1)
				}
				args := c.Args()
				client().Resolve(args[0])
				return nil
			},
		},
	}

	app.Run(os.Args)

}

func client() *lsrv.Client {
	return lsrv.NewClient(socket)
}
