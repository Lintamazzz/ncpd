package main

import (
	"fmt"
	"log"
	"ncpd/internal/auth"
)

func main() {
	token, err := auth.GetToken()
	if err != nil {
		log.Fatalf("获取 token 失败: %v", err)
	}

	videoID := "sm5zP5otVW5mZboEi9a8qQvJ"
	sessionID, err := auth.GetSessionID(videoID, token)
	if err != nil {
		log.Fatalf("获取 sessionID 失败: %v", err)
	}

	fmt.Printf("sessionID: %v\n", sessionID)
}
