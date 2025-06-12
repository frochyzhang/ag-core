package ip

import (
	"fmt"
	"log/slog"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// enableIPRange 的格式检查：,号分割多个配置，:号分割开始和结束ip段，.号分割的是ipv4的段，总ipv4的段数可以是1~4个，
// 案例：10.250.10:10.230,10.233、10.250
// var ipRangeRegex *regexp.Regexp = regexp.MustCompile(`^(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[1-9])(\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|\d)){0,3}(:(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[1-9])(\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|\d)){0,3}){0,1}(,(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[1-9])(\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|\d)){0,3}(:(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[1-9])(\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|\d)){0,3}){0,1})*$`)
var ipRangeRegex *regexp.Regexp = regexp.MustCompile(`^(((2[0-4]\d|25[0-5]|[01]?\d\d?)(\.(2[0-4]\d|25[0-5]|[01]?\d\d?)){0,3})(:((2[0-4]\d|25[0-5]|[01]?\d\d?)(\.(2[0-4]\d|25[0-5]|[01]?\d\d?)){0,3})){0,1}){1}(,(((2[0-4]\d|25[0-5]|[01]?\d\d?)(\.(2[0-4]\d|25[0-5]|[01]?\d\d?)){0,3})(:((2[0-4]\d|25[0-5]|[01]?\d\d?)(\.(2[0-4]\d|25[0-5]|[01]?\d\d?)){0,3})){0,1}){0,3})*$$`)

// ************** IPRanger **************
type IPRanger struct {
	IPRangeStr string
	IPRanges   []IPRange
}

func NewIPRanger(enableIPRange string) (*IPRanger, error) {
	ipr := &IPRanger{}
	ipr.IPRangeStr = enableIPRange

	if len(enableIPRange) != 0 && !ipRangeRegex.MatchString(enableIPRange) {
		// slog.Warn(fmt.Sprintf("[WARN]IPRange配置将不生效,非法的IPRange配置:%s\n", enableIPRange))
		// return nil, fmt.Errorf(" 非法的IPRange配置:%s\n", enableIPRange)
		slog.Warn(fmt.Sprintf("[WARN]invalid ip range: %s", enableIPRange))
		return nil, fmt.Errorf("invalid ip range: %s", enableIPRange)
	}

	iprs := make([]IPRange, 0)
	if len(enableIPRange) == 0 {
		iprs = append(iprs, NewIPRange("0", "255"))
	} else {
		ipRanges := strings.Split(enableIPRange, ",")
		for _, ipRange := range ipRanges {
			if len(ipRange) == 0 {
				continue
			}
			if strings.Contains(ipRange, ":") {
				ranges := strings.Split(ipRange, ":")
				iprs = append(iprs, NewIPRange(ranges[0], ranges[1]))
			} else {
				iprs = append(iprs, NewIPRange(ipRange, ipRange))
			}
		}
	}
	ipr.IPRanges = iprs

	return ipr, nil
}

func (r *IPRanger) GetLocalIP() (string, bool, error) {
	inters, err := net.Interfaces()
	if err != nil {
		return "", false, err
	}
	for _, inter := range inters {
		// 排除回环网口
		if inter.Flags&net.FlagUp != 0 && !strings.HasPrefix(inter.Name, "lo") {
			addrs, err := inter.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						// ip 范围可用判断
						if r.IPIsRangeAvailable(ipnet.IP.String()) {
							return ipnet.IP.String(), true, nil
						}
					}
				}
			}
		}
	}
	return "0.0.0.0", false, nil
}

// ************** IPRange **************
// IPRange ip范围
type IPRange struct {
	Start int64
	End   int64
}

func (r *IPRanger) IPIsRangeAvailable(ip string) bool {
	if len(ip) == 0 {
		return false
	}

	// IPRange 为空则返回true
	if r.IPRanges == nil || len(r.IPRanges) < 1 {
		return true
	}

	// 当前ip是否在配置的ip范围内
	for _, iprang := range r.IPRanges {
		if iprang.IsEnabled(ip) {
			return true
		}
	}

	return false
}

// NewIPRange 构建IPRange
func NewIPRange(startIP, endIP string) IPRange {
	return IPRange{
		Start: parseStart(startIP),
		End:   parseEnd(endIP),
	}
}

func parseStart(ip string) int64 {
	segments := []int{0, 0, 0, 0}
	return parse(segments, ip)
}

func parseEnd(ip string) int64 {
	segments := []int{255, 255, 255, 255}
	return parse(segments, ip)
}

func parse(segments []int, ip string) int64 {
	ipSegments := strings.Split(ip, ".")
	for i := 0; i < len(ipSegments); i++ {
		segments[i], _ = strconv.Atoi(ipSegments[i])
	}
	var ret int64
	for i := 0; i < len(segments); i++ {
		ret = ret*256 + int64(segments[i])
	}
	return ret
}

// IsEnabled 判断指定的ip是否在IP范围内
func (r IPRange) IsEnabled(ip string) bool {
	ipSegments := strings.Split(ip, ".")
	var ipInt int64
	for _, ipSegment := range ipSegments {
		val, _ := strconv.Atoi(ipSegment)
		ipInt = ipInt*256 + int64(val)
	}
	return ipInt >= r.Start && ipInt <= r.End
}
