package cvm

import (
	"strings"
	"fmt"
	"github.com/petermattis/goid"
)

func u2b(u1s []u1) []uint8 {
	bytes := make([]uint8, len(u1s))
	for i := 0; i < len(bytes); i++ {
		bytes[i] = uint8(u1s[i])
	}
	return bytes
}

func b2u(bytes []uint8) []u1 {
	bs := make([]u1, len(bytes))
	for i := 0; i < len(bytes); i++ {
		bs[i] = u1(bytes[i])
	}
	return bs
}

func u2s(u1s []u1) string {
	return string(u2b(u1s))
}

func u16toi32(i uint16) int32 {
	return int32(uint32(i))
}

func numberWithSign(i int32) string {
	if i >= 0 {
		return fmt.Sprintf("%s%d", "+", i)
	} else {
		return fmt.Sprintf("%s%d", "-", -i)
	}
}

func repeat(str string, times int) string {
	return strings.Repeat(str, times)
}

/*
A Java try {} catch() {} finally {} block
 */
type Block struct {
	try     func()
	catch   func(throwable Reference) // throwable never be nil
	finally func()
}

func (tcf Block) Do() {
	if tcf.finally != nil {
		defer tcf.finally()
	}
	if tcf.catch != nil {
		defer func() {
			if r := recover(); r != nil {
				if throwable, ok := r.(Reference); ok {
					tcf.catch(throwable)
				} else {
					// otherwise, the whole project has non-throwable panic,
					// But some 3rd-party package can panic other non-throwable
					Bug("CVM project has never non-throwable panic. "+
						"There is some 3rd-party package doing a non-throwable panic, check it. "+
						"Original panic: \n%v", r)
				}
			}
		}()
	}
	tcf.try()
}

func getGID() int64 {
	return goid.Get()
}

func binaryName2JavaName(name string) JavaLangString {
	return VM.NewJavaLangString(strings.Replace(name, "/", ".", -1))
}

func javaName2BinaryName(name JavaLangString) string {
	return JavaName2BinaryName0(name.ToNativeString())
}

func JavaName2BinaryName0(name string) string {
	return strings.Replace(name, ".", "/", -1)
}
