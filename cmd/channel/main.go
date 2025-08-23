package main

import (
	"fmt"
	"ncpd/internal/channel"
)

func main() {
	fmt.Println("获取频道列表:")
	channels, _ := channel.GetChannelList()
	for _, channel := range channels {
		fmt.Println(channel.Domain, channel.FanclubSite.ID)
	}
	fmt.Printf("总计 %d 个频道\n", len(channels))

	fmt.Printf("\n根据 ID 获取 域名: \n")
	ids := []int{38, 26, 25, 218, 387}
	for i, id := range ids {
		c1, _ := channel.GetChannelByID(id)
		fmt.Printf("%d -> %s\n", ids[i], c1.Domain)
	}

	fmt.Printf("\n根据 域名 获取 ID: \n")
	domain := "https://nicochannel.jp/sakakura-sakura"
	c2, _ := channel.GetChannelByDomain(domain)
	fmt.Printf("%s -> %d\n", domain, c2.FanclubSite.ID)
}
