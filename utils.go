package utils

import (
	"errors"
	"net"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func RemoveAnnotation(src []byte) []byte {
	reg := "(?P<nocomment>'(?:[^\\\\']|\\\\.)*'|\"(?:[^\\\\\"]|\\\\.)*\")|(?P<coment>//[^\n]*|/\\*(.|\n)*?\\*/)"
	re := regexp.MustCompile(reg)
	return re.ReplaceAll(src, []byte("${nocomment}"))
}

func IsInnerIp(src_ip string) bool {
	inet_network := func(ip string) uint32 {
		var (
			segments []string = strings.Split(ip, ".")
			ips      [4]uint64
			ret      uint64
		)
		for i := 0; i < 4; i++ {
			ips[i], _ = strconv.ParseUint(segments[i], 10, 64)
		}
		ret = ips[0]<<24 + ips[1]<<16 + ips[2]<<8 + ips[3]
		return uint32(ret)
	}

	ipa_beg := inet_network("10.0.0.0")
	ipa_end := inet_network("10.255.255.255")

	ipb_beg := inet_network("172.16.0.0")
	ipb_end := inet_network("172.31.255.255")

	ipc_beg := inet_network("192.168.0.0")
	ipc_end := inet_network("192.168.255.255")

	ip_seg := inet_network(src_ip)

	if (ip_seg >= ipa_beg && ip_seg <= ipa_end) || (ip_seg >= ipb_beg && ip_seg <= ipb_end) || (ip_seg >= ipc_beg && ip_seg <= ipc_end) {

		return true
	}

	return false
}

var GetLocalIp = func() func() string {
	localip := ""
	return func() string {
		if localip != "" {
			return localip
		}

		err := func() error {
			ifaces, err := net.Interfaces()
			if err != nil {
				return err
			}
			for _, iface := range ifaces {
				if iface.Flags&net.FlagUp == 0 {
					continue // interface down
				}
				if iface.Flags&net.FlagLoopback != 0 {
					continue // loopback interface
				}
				addrs, err := iface.Addrs()
				if err != nil {
					return err
				}
				for _, addr := range addrs {
					var ip net.IP
					switch v := addr.(type) {
					case *net.IPNet:
						ip = v.IP
					case *net.IPAddr:
						ip = v.IP
					}
					if ip == nil || ip.IsLoopback() {
						continue
					}
					ip = ip.To4()
					if ip == nil {
						continue // not an ipv4 address
					}
					if IsInnerIp(ip.String()) {
						localip = ip.String()
						return nil
					}
				}
			}

			return errors.New("GetLocalIp() fail")
		}()

		if err != nil {
			panic(err)
		}

		return localip
	}
}()

func Hash33(src string) int {
	hash := 5381
	for i := 0; i < len(src); i++ {
		hash = hash<<5 + hash + int(src[i])
	}
	return hash & 0x7fffffff
}

func Slice2Interface(slice interface{}) (ret []interface{}) {
	sliceRV := reflect.ValueOf(slice)
	num := sliceRV.Len()
	if num == 0 {
		return
	}
	ret = make([]interface{}, num)
	for pos := 0; pos < num; pos++ {
		ret[pos] = sliceRV.Index(pos).Interface()
	}
	return
}
