/*
 * Copyright (c) 2015-2018 Nexenta Systems, Inc.
 *
 * This file is part of EdgeFS Project
 * (see https://github.com/Nexenta/edgefs).
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */
package config

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/Nexenta/edgefs/src/efscli/efsutil"
	"github.com/spf13/cobra"
)

type RtlfsDevice struct {
	Name                string `json:"name"`
	Path                string `json:"path"`
	CheckMountpoint     int    `json:"check_mountpoint"`
	JournalMaxChunkSize int    `json:"journal_maxchunksize,omitempty"`
	Sync                int    `json:"sync"`
	Psize               int    `json:"psize,omitempty"`
	VerifyChid          int    `json:"verify_chid"`
	PlevelOverride      int    `json:"plevel_override,omitempty"`
	MaxSize             uint64 `json:"maxsize,omitempty"`
}

type RtlfsDevices struct {
	Devices []RtlfsDevice `json:"devices"`
}

type CcowTrlog struct {
	Interval   int `json:"interval,omitempty"`
	Quarantine int `json:"quarantine,omitempty"`
}

type CcowTenant struct {
	UnicastIO        int `json:"unicast_io"`
	FailureDomain    int `json:"failure_domain"`
	ReplicationCount int `json:"replication_count,omitempty"`
	SyncPut          int `json:"sync_put,omitempty"`
	SyncPutNamed     int `json:"sync_put_named,omitempty"`
}

type CcowNetwork struct {
	BrokerInterfaces string `json:"broker_interfaces"`
	ServerUnixSocket string `json:"server_unix_socket"`
	BrokerIP4addr    string `json:"broker_ip4addr,omitempty"`
	ServerIP4addr    string `json:"server_ip4addr,omitempty"`
	ServerPort       int    `json:"server_port,omitempty"`
}

type CcowConf struct {
	Trlog   CcowTrlog  `json:"trlog,omitempty"`
	Tenant  CcowTenant  `json:"tenant"`
	Network CcowNetwork `json:"network"`
}

type CcowdNetwork struct {
	ServerInterfaces string `json:"server_interfaces"`
	ServerUnixSocket string `json:"server_unix_socket"`
	ServerIP4addr    string `json:"server_ip4addr,omitempty"`
	ServerPort       int    `json:"server_port,omitempty"`
}

type CcowdBgConfig struct {
	TrlogDeleteAfterHours     int `json:"trlog_delete_after_hours,omitempty"`
	SpeculativeBackrefTimeout int `json:"speculative_backref_timeout,omitempty"`
}

type CcowdConf struct {
	BgConfig  CcowdBgConfig `json:"repdev_bg_config,omitempty"`
	Zone      int           `json:"zone,omitempty"`
	Network   CcowdNetwork  `json:"network"`
	Transport []string      `json:"transport"`
}

type AuditdConf struct {
	IsAggregator int `json:"is_aggregator"`
}

type NodeConfig struct {
	Ccow            CcowConf     `json:"ccow"`
	Ccowd           CcowdConf    `json:"ccowd"`
	Auditd          AuditdConf   `json:"auditd"`
	IPv4Autodetect  int          `json:"ipv4_autodetect,omitempty"`
	RtlfsAutodetect string       `json:"rtlfs_autodetect,omitempty"`
	ClusterNodes    []string     `json:"cluster_nodes,omitempty"`
	Rtrd            RTDevices    `json:"rtrd"`
	RtrdSlaves      []RTDevices  `json:"rtrdslaves"`
	Rtlfs           RtlfsDevices `json:"rtlfs"`
	NodeType        string       `json:"nodeType"`
}

var (
	/* default pathes */
	DefaultNedgePrefix = "/opt/nedge"
	CCOWJsonFile       = "/etc/ccow/ccow.json"
	CCOWDJsonFile      = "/etc/ccow/ccowd.json"
	RTRDJsonFile       = "/etc/ccow/rt-rd.json"
	RTLFSJsonFile      = "/etc/ccow/rt-lfs.json"

	CorosyncConfFile            = "/etc/corosync/corosync.conf"
	CorosyncConfExampleFile     = "/etc/corosync/corosync.conf.example"
	CorosyncConfIPv4ExampleFile = "/etc/corosync/corosync.conf.example.ipv4"

	FlexHashCheckpointFile = "/var/run/flexhash-checkpoint.json"

	AuditdIniFile        = "/etc/ccow/auditd.ini"
	AuditdIniFileExample = "/etc/ccow/auditd.ini.example"

	CcowdUnixSocketFile = "/var/run/sock/ccowd.sock"

	/* Loaded instance */
	nodeConfig    *NodeConfig
	clusterConfig map[string]*NodeConfig

	nedgeHome string
)

func ConfigNode() {

	fmt.Println("")

	if nodeConfig == nil {
		fmt.Printf("Node configuration is missing\n")
		os.Exit(1)
	}

	// Hardcode UnicastIO value to prevent multicast
	nodeConfig.Ccow.Tenant.UnicastIO = 3

	if nodeConfig.IPv4Autodetect == 1 {
		serverIP, err := efsutil.GetIPv4Address(nodeConfig.Ccowd.Network.ServerInterfaces)
		if err != nil {
			fmt.Printf("Can't find IP accesible address via network interface %s Error: %v \n", nodeConfig.Ccowd.Network.ServerInterfaces, err)
			os.Exit(1)
		}

		nodeConfig.Ccowd.Network.ServerIP4addr = serverIP
		nodeConfig.Ccow.Network.ServerIP4addr = serverIP

		brokerIP, err := efsutil.GetIPv4Address(nodeConfig.Ccow.Network.BrokerInterfaces)
		if err != nil {
			fmt.Printf("Can't find IP accesible address via network interface %s Error: %v \n", nodeConfig.Ccow.Network.BrokerInterfaces, err)
			os.Exit(1)
		}

		nodeConfig.Ccow.Network.BrokerIP4addr = brokerIP

		// corosync.conf
		err = efsutil.CopyFile(nedgeHome+CorosyncConfIPv4ExampleFile, nedgeHome+CorosyncConfFile)
		if err != nil {
			fmt.Printf("Can't copy corosync file %s to %s Error: %v \n", nedgeHome+CorosyncConfIPv4ExampleFile, nedgeHome+CorosyncConfFile, err)
			os.Exit(1)
		}

		var nodenames []string
		if nodeConfig.ClusterNodes != nil {
			nodenames = make([]string, 0, len(nodeConfig.ClusterNodes))
			for i := range nodeConfig.ClusterNodes {
				nodenames = append(nodenames, nodeConfig.ClusterNodes[i])
			}
		} else {
			nodenames = make([]string, 0, len(clusterConfig))
			for nn := range clusterConfig {
				nodenames = append(nodenames, nn)
			}
		}

		var nodeList strings.Builder
		nodeList.WriteString("\nnodelist {\n")

		for i := range nodenames {
			nodeFQDN := nodenames[i]
			node := strings.Split(nodeFQDN, ".")[0]

			var cname string
			cname, err = net.LookupCNAME(nodeFQDN)
			if err != nil {
				cname = ""
			}

			var nodeName string

			// lookup by CNAME first
			if cname != "" {
				nodeName = cname
				_, err = efsutil.LookupDNS(cname)
				if err != nil {
					// no luck, lookup by supplied FQDN
					nodeName = nodeFQDN
					_, err = efsutil.LookupDNS(nodeFQDN)
				}
			} else {
				nodeName = nodeFQDN
				_, err = efsutil.LookupDNS(nodeFQDN)
			}
			if err != nil {
				// no luck, lookup by hostname
				nodeName = node
				_, err = efsutil.LookupDNS(node)
				if err != nil {
					fmt.Printf("Can't lookup DNS entry %s\n", nodeFQDN)
					os.Exit(1)
				}
			}
			nodeIdUint32 := efsutil.GetMD5HashInt32(nodeFQDN) / 2

			nodeList.WriteString("  node {\n")
			nodeList.WriteString("    ring0_addr: " + nodeName + "\n")
			nodeList.WriteString("    nodeid: " + fmt.Sprint(nodeIdUint32) + "\n")
			nodeList.WriteString("  }\n")
		}
		nodeList.WriteString("}\n")
		//fmt.Printf(" Nodelist %s\n", nodeList.String())
		err = efsutil.AppendStringToFile(nedgeHome+CorosyncConfFile, nodeList.String())
		if err != nil {
			fmt.Printf("Can't appent corosync file %s  Error: %v \n", nedgeHome+CorosyncConfFile, err)
			os.Exit(1)
		}

		input, err := ioutil.ReadFile(nedgeHome + CorosyncConfFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// adjust log file location
		output := regexp.MustCompile(`/opt/nedge`).ReplaceAllString(string(input), os.Getenv("NEDGE_HOME"))

		// adjust netmtu to the current value of selected server interface name
		ifname0 := strings.Split(nodeConfig.Ccowd.Network.ServerInterfaces, ";")[0]
		netmtu := detectMTU(ifname0)
		output = regexp.MustCompile(`netmtu:.*`).ReplaceAllString(output, "netmtu: "+strconv.Itoa(netmtu))

		if err = ioutil.WriteFile(nedgeHome+CorosyncConfFile, []byte(output), 0666); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("Configured %s to use IPv4 and explicit node list, netmtu=%v\n",
			nedgeHome+CorosyncConfFile, netmtu)
	} else {
		// IPv6 and multicast

		// corosync.conf
		err := efsutil.CopyFile(nedgeHome+CorosyncConfExampleFile, nedgeHome+CorosyncConfFile)
		if err != nil {
			fmt.Printf("Can't copy corosync file %s to %s Error: %v \n", nedgeHome+CorosyncConfIPv4ExampleFile, nedgeHome+CorosyncConfFile, err)
			os.Exit(1)
		}

		input, err := ioutil.ReadFile(nedgeHome + CorosyncConfExampleFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		ifname0 := strings.Split(nodeConfig.Ccowd.Network.ServerInterfaces, ";")[0]

		// adjust netmtu to the current value of selected server interface name
		netmtu := detectMTU(ifname0)
		output := regexp.MustCompile(`netmtu:.*`).ReplaceAllString(string(input), "netmtu: "+strconv.Itoa(netmtu))

		// adjust nodeid, needed for IPv6 and multicase case
		ether := detectEther(ifname0)
		nodeIdUint32 := efsutil.GetMD5HashInt32(ether) / 2
		output = regexp.MustCompile(`nodeid:.*`).ReplaceAllString(output, "nodeid: "+strconv.Itoa(int(nodeIdUint32)))

		// adjust bindnetaddr to the selected server interface name
		output = regexp.MustCompile(`bindnetaddr:.*`).ReplaceAllString(output, "bindnetaddr: "+ifname0)

		// adjust log file location
		output = regexp.MustCompile(`/opt/nedge`).ReplaceAllString(output, os.Getenv("NEDGE_HOME"))

		if err = ioutil.WriteFile(nedgeHome+CorosyncConfFile, []byte(output), 0666); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Configured %s to use IPv6 and autodetect, nodeid=%v, netmtu=%v\n",
			nedgeHome+CorosyncConfFile, nodeIdUint32, netmtu)
	}

	if &nodeConfig.Ccow != nil {
		err := efsutil.MarshalToFile(nedgeHome+CCOWJsonFile, &nodeConfig.Ccow)
		if err != nil {
			fmt.Printf("Can't marshal JSON file %s Error: %v \n", nedgeHome+CCOWJsonFile, err)
			os.Exit(1)
		}
		fmt.Printf("Configured %s\n", nedgeHome+CCOWJsonFile)
	}

	transport := "rtlfs"
	if &nodeConfig.Ccowd != nil {
		err := efsutil.MarshalToFile(nedgeHome+CCOWDJsonFile, &nodeConfig.Ccowd)
		if err != nil {
			fmt.Printf("Can't marshal JSON file %s Error: %v \n", nedgeHome+CCOWDJsonFile, err)
			os.Exit(1)
		}
		transport = nodeConfig.Ccowd.Transport[0]
		fmt.Printf("Configured %s, transport=%s\n", nedgeHome+CCOWDJsonFile, transport)
	}

	if transport == "rtrd" {
		// we should generate empty rtrd file in case of nodeType=="gateway" only
		// if there are devices list without gateway nodetype it is resurrect case and no need to generate rt-rd file
		fmt.Printf("RTRD configuration for NodeType: %s %+v", nodeConfig.NodeType, nodeConfig.Rtrd)
		if len(nodeConfig.Rtrd.Devices) > 0 || nodeConfig.NodeType == "gateway" {
			err := efsutil.MarshalToFile(nedgeHome+RTRDJsonFile, &nodeConfig.Rtrd)
			if err != nil {
				fmt.Printf("Can't marshal JSON file %s Error: %v \n", nedgeHome+RTRDJsonFile, err)
				os.Exit(1)
			}

			fmt.Printf("Configured master daemon %s\n", nedgeHome+RTRDJsonFile)
		} else {
			fmt.Printf("Warning: No RTRD devices passed. No rt-rd.json created.\n")
		}
	} else if len(nodeConfig.RtlfsAutodetect) > 0 {
		rootDir := nodeConfig.RtlfsAutodetect
		if !efsutil.IsDirectory(rootDir) {
			fmt.Printf("Rt-lfs autodetect points to non existing object or to file\n")
			os.Exit(1)
		}

		devices, err := getRtlfsDevices(rootDir)
		if err != nil {
			fmt.Printf("Can't get rtlfs devices %s Error: %v \n", nedgeHome+RTLFSJsonFile, err)
			os.Exit(1)
		}

		if len(devices) == 0 {
			//create devices, should be passed as parameter
			for i := 0; i < 4; i++ {
				CreateDirIfNotExist(rootDir + "/" + "device-" + fmt.Sprint(i))
			}
			devices, err = getRtlfsDevices(rootDir)
			if err != nil {
				fmt.Printf("Can't get rtlfs devices %s Error: %v \n", nedgeHome+RTLFSJsonFile, err)
				os.Exit(1)
			}
		}

		rtlfsJSONObj := &RtlfsDevices{
			Devices: devices,
		}
		err = efsutil.MarshalToFile(nedgeHome+RTLFSJsonFile, rtlfsJSONObj)
		if err != nil {
			fmt.Printf("Can't marshal JSON file %s Error: %v \n", nedgeHome+RTLFSJsonFile, err)
			os.Exit(1)
		}

		fmt.Printf("Configured %s\n", nedgeHome+RTLFSJsonFile)

	} else if transport == "rtlfs" {

		// we should generate empty rtlfs file in case of nodeType=="gateway" only
		// if there are devices list without gateway nodetype it is resurrect case and no need to generate rt-rd file
		fmt.Printf("Rtlfs configuration for NodeType: %s %+v", nodeConfig.NodeType, nodeConfig.Rtlfs)
		if len(nodeConfig.Rtlfs.Devices) > 0 || nodeConfig.NodeType == "gateway" {
			err := efsutil.MarshalToFile(nedgeHome+RTLFSJsonFile, &nodeConfig.Rtlfs)
			if err != nil {
				fmt.Printf("Can't marshal JSON file %s Error: %v \n", nedgeHome+RTLFSJsonFile, err)
				os.Exit(1)
			}
			fmt.Printf("Configured %s\n", nedgeHome+RTLFSJsonFile)
		} else {
			fmt.Printf("Warning: No Rtlfs devices passed. No %s created.\n", nedgeHome+RTLFSJsonFile)
		}

	} else {
		fmt.Printf("Driver transport not re-configured\n")
	}

	/* auditd.ini */
	efsutil.CopyFile(nedgeHome+AuditdIniFileExample, nedgeHome+AuditdIniFile)
	if nodeConfig.Auditd.IsAggregator == 1 {
		efsutil.ReplaceInFile(nedgeHome+AuditdIniFile, `is_aggregator=\d`, "is_aggregator="+strconv.Itoa(nodeConfig.Auditd.IsAggregator))
	}
	efsutil.ReplaceInFile(nedgeHome+AuditdIniFile, `/opt/nedge`, os.Getenv("NEDGE_HOME"))
	fmt.Printf("Configured %s\n", nedgeHome+AuditdIniFile)

	fmt.Printf("\nConfiguration applied successfully.\n\n")
}

func getRtlfsDevices(rootDir string) (devices []RtlfsDevice, err error) {
	devices = make([]RtlfsDevice, 0)
	folderEntries, err := efsutil.GetFolderContent(rootDir)
	if err != nil {
		fmt.Printf("Rt-lfs can't get entire folder objects of %s. %v\n", rootDir, err)
		return nil, err
	}

	for _, entry := range folderEntries {
		if entry.IsDir() {
			//skip .state subfolder in data dir
			if entry.Name() == ".state" || entry.Name() == ".etc" {
				continue
			}

			path := rootDir + "/" + entry.Name()
			device := RtlfsDevice{
				Name:            entry.Name(),
				Path:            path,
				CheckMountpoint: 0,
				Sync:            1,
			}

			devices = append(devices, device)
		}
	}

	return devices, nil
}

func CreateDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func detectMTU(ifname string) int {
	cmd := "ip link show " + ifname + " | grep ' mtu ' | sed -e 's/.* mtu \\([0-9]\\+\\).*/\\1/'"
	mtu, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil || string(mtu) == "" {
		mtu = []byte("1500")
	}

	netmtu, err := strconv.Atoi(strings.TrimSpace(string(mtu)))
	if err != nil {
		netmtu = 1500
	}

	return netmtu - 50
}

func detectEther(ifname string) string {
	cmd := "ip link show " + ifname + " | awk '/ether/{print $2}'"
	ether, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil || string(ether) == "" {
		fmt.Printf("Warning: cannot detect Ether address of %s\n", ifname)
		ether = []byte("a:b:c:d:e:f")
	}
	return strings.TrimSpace(string(ether))
}

var (
	name string

	ConfigCmd = &cobra.Command{
		Use:     "config",
		Aliases: []string{"c"},
		Short:   "Node configuration operations",
		Long:    "Node configuration operations, e.g. via file, automatic, profile-based",
	}
)

func init() {
}
