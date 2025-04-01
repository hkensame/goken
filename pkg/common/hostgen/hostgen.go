package hostgen

import (
	"net"
	"strconv"
	"strings"

	"github.com/hkensame/goken/pkg/errors"
	"github.com/hkensame/goken/pkg/log"
)

var ErrInvalidHost = errors.New("错误的host格式")

func isValidIPAndLocalHost(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		if ip == "localhost" || ip == "127.0.0.1" {
			return true
		}
		return false
	}

	return parsedIP.To4() != nil || parsedIP.To16() != nil
}

// func isValidIPButLocalHost(addr string) bool {
// 	ip := net.ParseIP(addr)
// 	return ip.IsGlobalUnicast() && !ip.IsInterfaceLocalMulticast()
// }

func isValidPort(port any) bool {
	switch t := port.(type) {
	case string:
		p, err := strconv.Atoi(t)
		if err != nil {
			return false
		}
		return p >= 1 && p <= 65535
	case int, int32, int64:
		p := port.(int)
		return p >= 1 && p <= 65535
	}
	return false
}

func GetUsagePort() int {
	//把字符串解析为tcp端点
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	//tcp协议中,如果端口为0则在listen,dial这些函数中会默认给它分配一个空闲的端口
	lis, _ := net.ListenTCP("tcp", addr)
	defer lis.Close()
	return lis.Addr().(*net.TCPAddr).Port

}

func ResolveHost(host string) (string, error) {
	ip, port, err := net.SplitHostPort(host)
	if err != nil {
		log.Errorf("[endpoint] 无法从 host 中提取有效的 IP 或端口")
		return "", ErrInvalidHost
	}

	if uport, err := strconv.Atoi(port); uport <= 0 || err != nil {
		uport = GetUsagePort()
		port = strconv.Itoa(uport)
	}

	if len(ip) > 0 && ip != "0.0.0.0" && ip != "::" && ip != "[::]" {
		return net.JoinHostPort(ip, port), nil
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	var preferredIP, fallbackIP net.IP
	for _, iface := range ifaces {
		if (iface.Flags&net.FlagUp) == 0 || (iface.Flags&net.FlagLoopback) != 0 {
			continue
		}

		if strings.HasPrefix(iface.Name, "docker") || strings.HasPrefix(iface.Name, "veth") {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, rawAddr := range addrs {
			var ip net.IP
			switch addr := rawAddr.(type) {
			case *net.IPAddr:
				ip = addr.IP
			case *net.IPNet:
				ip = addr.IP
			default:
				continue
			}

			if ip.To4() != nil && !ip.IsLoopback() {
				if iface.Name == "ens33" || iface.Name == "eth0" || strings.HasPrefix(iface.Name, "enp") {
					preferredIP = ip
					break
				} else if fallbackIP == nil {
					fallbackIP = ip
				}
			}
		}
	}

	if preferredIP != nil {
		return net.JoinHostPort(preferredIP.String(), port), nil
	}
	if fallbackIP != nil {
		return net.JoinHostPort(fallbackIP.String(), port), nil
	}

	return "", nil
}

func ValidListenHost(host string) bool {
	ip, port, err := net.SplitHostPort(host)
	if err != nil {
		return false
	}
	return isValidIPAndLocalHost(ip) && isValidPort(port)
}
