package main

import (
	"fmt"
	"ncpd/internal/auth"
	"ncpd/internal/news"
)

func main() {
	fcSiteID := 387
	articles, err := news.GetArticleList(fcSiteID)
	if err != nil {
		fmt.Println("获取文章列表失败:", err)
		return
	}

	token, err := auth.GetToken()
	if err != nil {
		fmt.Println("获取Token失败:", err)
		return
	}

	for i, a := range articles {
		article, err := news.GetArticle(fcSiteID, a.ArticleCode, token)
		if err != nil {
			fmt.Println("获取文章失败:", err)
			continue
		}

		fmt.Printf("\n\n %d. %s\n", i+1, article.ArticelTitle)
		fmt.Printf("  ID: %d\n", article.ID)
		fmt.Printf("  编号: %s\n", a.ArticleCode)
		fmt.Printf("  封面: %s\n", article.ThumbnailURL)
		fmt.Printf("  时间: %s\n", article.PublishAt)
		// fmt.Printf("  内容: %s\n", article.Contents)
		if article.Contents == "" {
			fmt.Println(" ❌ 内容: 无")
		} else {
			fmt.Println(" ✅ 内容: 有")
		}
	}

}
