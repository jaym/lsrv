package main

import (
	"log"
	"net"
	"os"

	"github.com/jaym/lsrv"
	cli "gopkg.in/urfave/cli.v1"
	"gopkg.in/urfave/cli.v1/altsrc"
)

func main() {
	app := cli.NewApp()
	app.Name = "lsrv"

	flags := []cli.Flag{
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "ip_block",
			Value: "172.22.0.0/24",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "state_file",
			Value: "./state_file",
		}),
		cli.StringFlag{
			Name:  "config, c",
			Value: "/etc/lsrv.toml",
		},
	}

	app.Before = func(c *cli.Context) error {
		f := altsrc.InitInputSourceWithContext(flags, altsrc.NewTomlSourceFromFlagFunc("config"))
		f(c)
		return nil
	}
	app.Flags = flags

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
				client(c).Add(args[0], "127.0.0.1", args[1], args[2])
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
				client(c).Delete(args[0])
				return nil
			},
		},
		{
			Name:        "restore",
			Usage:       "Restore all services",
			Description: "Populates iptables and the hosts file based on the current state",
			Action: func(c *cli.Context) error {
				if len(c.Args()) != 0 {
					cli.ShowCommandHelpAndExit(c, "restore", 1)
				}
				client(c).Restore()
				return nil
			},
		},
		{
			Name:        "cleanup",
			Usage:       "Remove all services from iptables and the hosts file",
			Description: "Remove all services from iptables and the hosts file",
			Action: func(c *cli.Context) error {
				if len(c.Args()) != 0 {
					cli.ShowCommandHelpAndExit(c, "cleanup", 1)
				}
				client(c).Cleanup()
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
				client(c).Resolve(args[0])
				return nil
			},
		},
	}

	app.Run(os.Args)

}

func client(c *cli.Context) *lsrv.Client {
	_, ip_block, err := net.ParseCIDR(c.Parent().String("ip_block"))
	if err != nil {
		log.Fatal("Invalid ip_block: ", err)
	}
	return lsrv.NewClient(c.Parent().String("state_file"), ip_block)
}
