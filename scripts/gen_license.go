package main

import (
	"crypto/md5"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run scripts/gen_license.go [特征码]")
		return
	}

	machineID := strings.ToUpper(os.Args[1])
	// 必须与 internal/config/license.go 中的盐值完全一致
	salt := "legal-extractor-secret-2026"
	
	raw := fmt.Sprintf("%x", md5.Sum([]byte(machineID+salt)))
	code := strings.ToUpper(raw[:16])
	license := fmt.Sprintf("%s-%s-%s-%s", code[0:4], code[4:8], code[8:12], code[12:16])

	fmt.Printf("\n==================================\n")
	fmt.Printf("特征码: %s\n", machineID)
	fmt.Printf("授权码: %s\n", license)
	fmt.Printf("==================================\n")
}
