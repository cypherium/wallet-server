package nat

import (
	"fmt"
	"net"
	"strings"
)

func GetVirtualIp(namePrefix string) string {
	ip := ""
	intfs, err := net.Interfaces()
	if err != nil {
		fmt.Printf("Get Addr error:\n%s\n", err)
		return ""
	}

	_, LanIP10, _ := net.ParseCIDR("10.0.0.0/8")
	_, LanIP172, _ := net.ParseCIDR("172.18.0.0/16")
	_, LanIP192, _ := net.ParseCIDR("192.168.0.0/16")

	for _, intf := range intfs {
		list, _ := intf.Addrs()
		for _, addr := range list {
			if !strings.Contains(addr.String(), ":") {
				list := strings.Split(addr.String(), "/")
				if len(list) != 2 {
					continue
				}
				i := net.ParseIP(list[0])
				if i == nil {
					continue
				}

				if LanIP10.Contains(i) || LanIP172.Contains(i) || LanIP192.Contains(i) {
					if strings.Contains(intf.Name, namePrefix) {
						ip = i.String()
						goto out
					}
				}

			}
		}
	}
out:
	return ip
}
