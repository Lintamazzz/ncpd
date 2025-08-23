package main

import (
	"fmt"
	"log"

	auth "ncpd/internal/auth"
)

func main() {
	refresh_token := auth.GetRefreshToken()
	fmt.Printf("请求前的 refresh_token: %s\n\n", refresh_token)

	token, err := auth.GetToken()
	if err != nil {
		log.Fatalf("获取 token 失败: %v", err)
	}

	fmt.Println(token)

	refresh_token = auth.GetRefreshToken()
	fmt.Println("\n请求后的 refresh_token: ", refresh_token)
}
