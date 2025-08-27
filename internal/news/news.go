package news

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

type Article struct {
	ID           int    `json:"id"`
	ArticleCode  string `json:"article_code"`
	ArticelTitle string `json:"article_title"`
	Contents     string `json:"contents"`
	PublishAt    string `json:"publish_at"`
	ThumbnailURL string `json:"thumbnail_url"`
}

type ArticleResponse struct {
	Data struct {
		Article struct {
			Article *Article `json:"article"`
		} `json:"article"`
	} `json:"data"`
}

// ArticlesResponse 文章列表响应结构体
type ArticlesResponse struct {
	Data struct {
		ArticleTheme struct {
			Articles struct {
				List  []Article `json:"list"`
				Total int       `json:"total"`
			} `json:"articles"`
		} `json:"article_theme"`
	} `json:"data"`
}

// 返回的 article.contents 不是原文，原文需要用 GetArticle 获取
func GetArticleList(fcSiteID int) ([]Article, error) {
	client := resty.New()
	page := 1
	size := 24
	baseURL := "https://api.nicochannel.jp/fc/fanclub_sites/%d/article_themes/news/articles?per_page=%d&sort=published_at_desc&page=%d"

	var allArticles []Article

	fmt.Printf("开始获取 频道ID: %d 的所有文章信息...\n", fcSiteID)

	for {
		var articlesResponse ArticlesResponse
		url := fmt.Sprintf(baseURL, fcSiteID, size, page)

		resp, err := client.R().
			SetHeader("fc_use_device", "null").
			SetResult(&articlesResponse).
			Get(url)

		if err != nil {
			return nil, fmt.Errorf("GetArticleList: 请求第 %d 页失败 %w", page, err)
		}

		// 打印响应状态码
		fmt.Printf("第 %d 页响应状态码: %d\n", page, resp.StatusCode())

		// 检查是否有数据
		if len(articlesResponse.Data.ArticleTheme.Articles.List) == 0 {
			fmt.Printf("第 %d 页没有数据，停止获取\n\n", page)
			break
		}

		// 将当前页的数据添加到总列表中
		allArticles = append(allArticles, articlesResponse.Data.ArticleTheme.Articles.List...)

		fmt.Printf("第 %d 页获取到 %d 篇文章，累计 %d 篇文章\n",
			page, len(articlesResponse.Data.ArticleTheme.Articles.List), len(allArticles))

		// 如果当前页的数据少于每页数量，说明已经是最后一页
		if len(articlesResponse.Data.ArticleTheme.Articles.List) < size {
			fmt.Printf("第 %d 页数据少于 %d 个，已到达最后一页\n", page, size)
			break
		}

		page++
	}

	return allArticles, nil
}

// 要带 token，不然会员内容 contents 会返回空
func GetArticle(fcSiteID int, articleCode string, token string) (*Article, error) {
	client := resty.New()

	baseURL := "https://api.nicochannel.jp/fc/fanclub_sites/%d/article_themes/news/articles/%s"
	url := fmt.Sprintf(baseURL, fcSiteID, articleCode)

	var articleResponse ArticleResponse

	resp, err := client.R().
		SetHeader("fc_use_device", "null").
		SetAuthToken(token).
		SetResult(&articleResponse).
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("GetArticle: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("状态码 %d", resp.StatusCode())
	}

	return articleResponse.Data.Article.Article, nil
}
