package Common

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

var (
	mu 					sync.Mutex
	baseUserID			uint64
)

func ToString(i int64) string {
	str := fmt.Sprintf("%d", i)
	return str
}

func ToInt64(str string) int64 {
	i,err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return int64(i)
}

func KeyInput() string {
	InputBytes := make([]byte, 512)
	_,err := os.Stdin.Read(InputBytes)
	if err != nil {
		fmt.Println("read error:", err)
	}
	textBytes := bytes.TrimRight(InputBytes, "\x00")
	in := strings.Trim(strings.Trim(string(textBytes),"\n")," ")
	return in
}

func StrToBytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	b := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&b))
}

func GetStructName(i interface{}) string {
	PdName := reflect.TypeOf(i).String()
	strs := strings.Split(PdName,".")
	return strs[len(strs) - 1]
}

func GetNow() int64 {
	return time.Now().Unix()
}

func SetUserID(iniUserID uint64) {
	mu.Lock()
	baseUserID = iniUserID
	mu.Unlock()
}

//线程安全的字增序列
func GenUserID() uint64 {
	mu.Lock()
	baseUserID += 1
	mu.Unlock()
	return baseUserID
}

func MaxUInt64(a uint64,b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}