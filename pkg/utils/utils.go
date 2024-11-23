package utils

import (
	"crypto/rand"
	"errors"
	"math/big"
	"net/netip"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/exp/constraints"

	jsoniter "github.com/json-iterator/go"
)

var (
	Json = jsoniter.ConfigCompatibleWithStandardLibrary

	DNSServers = []string{"1.1.1.1:53", "223.5.5.5:53"}
)

var ipv4Re = regexp.MustCompile(`(\d*\.).*(\.\d*)`)

func ipv4Desensitize(ipv4Addr string) string {
	return ipv4Re.ReplaceAllString(ipv4Addr, "$1****$2")
}

var ipv6Re = regexp.MustCompile(`(\w*:\w*:).*(:\w*:\w*)`)

func ipv6Desensitize(ipv6Addr string) string {
	return ipv6Re.ReplaceAllString(ipv6Addr, "$1****$2")
}

func IPDesensitize(ipAddr string) string {
	ipAddr = ipv4Desensitize(ipAddr)
	ipAddr = ipv6Desensitize(ipAddr)
	return ipAddr
}

func IPStringToBinary(ip string) ([]byte, error) {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return nil, err
	}
	b := addr.As16()
	return b[:], nil
}

func BinaryToIPString(b []byte) string {
	var addr16 [16]byte
	copy(addr16[:], b)
	addr := netip.AddrFrom16(addr16)
	return addr.Unmap().String()
}

func GetIPFromHeader(headerValue string) (string, error) {
	a := strings.Split(headerValue, ",")
	h := strings.TrimSpace(a[len(a)-1])
	ip, err := netip.ParseAddrPort(h)
	if err != nil {
		return "", err
	}
	if !ip.IsValid() {
		return "", errors.New("invalid ip")
	}
	return ip.Addr().String(), nil
}

// SplitIPAddr 传入/分割的v4v6混合地址，返回v4和v6地址与有效地址
func SplitIPAddr(v4v6Bundle string) (string, string, string) {
	ipList := strings.Split(v4v6Bundle, "/")
	ipv4 := ""
	ipv6 := ""
	validIP := ""
	if len(ipList) > 1 {
		// 双栈
		ipv4 = ipList[0]
		ipv6 = ipList[1]
		validIP = ipv4
	} else if len(ipList) == 1 {
		// 仅ipv4|ipv6
		if strings.Contains(ipList[0], ":") {
			ipv6 = ipList[0]
			validIP = ipv6
		} else {
			ipv4 = ipList[0]
			validIP = ipv4
		}
	}
	return ipv4, ipv6, validIP
}

func IsFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	lettersLength := big.NewInt(int64(len(letters)))
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, lettersLength)
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}
	return string(ret), nil
}

func Uint64SubInt64(a uint64, b int64) uint64 {
	if b < 0 {
		return a + uint64(-b)
	}
	if a < uint64(b) {
		return 0
	}
	return a - uint64(b)
}

func IfOr[T any](a bool, x, y T) T {
	if a {
		return x
	}
	return y
}

func IfOrFn[T any](a bool, x, y func() T) T {
	if a {
		return x()
	}
	return y()
}

func Itoa[T constraints.Integer](i T) string {
	switch any(i).(type) {
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(int64(i), 10)
	case uint, uint8, uint16, uint32, uint64:
		return strconv.FormatUint(uint64(i), 10)
	default:
		return ""
	}
}

// From go1.23

// Compare returns
//
//	-1 if x is less than y,
//	 0 if x equals y,
//	+1 if x is greater than y.
//
// For floating-point types, a NaN is considered less than any non-NaN,
// a NaN is considered equal to a NaN, and -0.0 is equal to 0.0.
func Compare[T constraints.Ordered](x, y T) int {
	xNaN := isNaN(x)
	yNaN := isNaN(y)
	if xNaN {
		if yNaN {
			return 0
		}
		return -1
	}
	if yNaN {
		return +1
	}
	if x < y {
		return -1
	}
	if x > y {
		return +1
	}
	return 0
}

// isNaN reports whether x is a NaN without requiring the math package.
// This will always return false if T is not floating-point.
func isNaN[T constraints.Ordered](x T) bool {
	return x != x
}
