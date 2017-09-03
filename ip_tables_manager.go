package lsrv

import (
	"log"
	"strconv"

	"github.com/coreos/go-iptables/iptables"
)

type IPTablesManager struct {
	ipt *iptables.IPTables
}

func NewIPTablesManager() (*IPTablesManager, error) {
	ipt, err := iptables.New()
	if err != nil {
		return nil, err
	} else {
		manager := new(IPTablesManager)
		manager.ipt = ipt
		return manager, nil
	}
}

func (manager *IPTablesManager) Initialize() error {
	ipt := manager.ipt

	ipt.NewChain("nat", "LSRV")
	ipt.AppendUnique("nat", "OUTPUT", "-jLSRV")

	return nil
}

func (manager *IPTablesManager) AddRule(source_addr string, source_port uint16,
	dest_addr string, dest_port uint16) error {

	ipt := manager.ipt
	rulespec := rule_for(source_addr, source_port, dest_addr, dest_port)

	ipt.AppendUnique("nat", "LSRV", rulespec...)

	return nil
}

func (manager *IPTablesManager) RemoveRule(source_addr string, source_port uint16,
	dest_addr string, dest_port uint16) error {

	ipt := manager.ipt
	rulespec := rule_for(source_addr, source_port, dest_addr, dest_port)

	return ipt.Delete("nat", "LSRV", rulespec...)
}

func (manager *IPTablesManager) Cleanup() error {
	ipt := manager.ipt

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

func rule_for(service_addr string, service_port uint16,
	dest_addr string, dest_port uint16) []string {

	return []string{"-p", "tcp", "-d", dest_addr, "--dport",
		strconv.FormatUint(uint64(dest_port), 10), "-j", "DNAT",
		"--to", service_addr + ":" + strconv.FormatUint(uint64(service_port), 10)}
}
