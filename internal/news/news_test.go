package news

import (
	"testing"

	"ncpd/internal/auth"
)

// go test -v ./internal/news
func TestNewsAPI(t *testing.T) {
	fcSiteID := 387

	// 1. 获取文章列表
	t.Log("🔍 正在获取文章列表...")
	articles, err := GetArticleList(fcSiteID)
	if err != nil {
		t.Fatalf("获取文章列表失败: %v", err)
	}

	if len(articles) == 0 {
		t.Log("获取到的文章列表为空")
		return
	}

	t.Logf("✅ 成功获取到 %d 篇文章", len(articles))

	// 2. 输出每篇文章的基本信息
	t.Log("📋 文章列表:")
	for i, article := range articles {
		t.Logf("  %d. [%s] %s", i+1, article.ArticleCode, article.ArticelTitle)

		// 验证文章结构
		if article.ArticleCode == "" {
			t.Errorf("文章 %d 的 ArticleCode 为空", i+1)
		}
		if article.ArticelTitle == "" {
			t.Errorf("文章 %d 的标题为空", i+1)
		}
	}

	// 3. 获取第一篇文章的详细信息
	t.Log("📄 正在获取第一篇文章的详细信息...")

	// 获取 token
	token, err := auth.GetToken()
	if err != nil {
		t.Fatal("获取Token失败:", err)
	}

	// 获取第一篇文章的详细信息
	article, err := GetArticle(fcSiteID, articles[0].ArticleCode, token)
	if err != nil {
		t.Fatalf("获取文章详情失败: %v", err)
	}

	// 验证文章详情
	if article.ArticelTitle == "" {
		t.Error("文章标题不应为空")
	}

	if article.Contents == "" {
		t.Error("文章内容不应为空")
	}

	// 4. 输出文章详情信息
	t.Logf("✅ 成功获取文章详情:")
	t.Logf("  标题: %s", article.ArticelTitle)
	t.Logf("  发布时间: %s", article.PublishAt)
	t.Logf("  内容长度: %d 字符", len(article.Contents))
	if len(article.Contents) > 100 {
		t.Logf("  内容预览: %s...", article.Contents[:100])
	} else {
		t.Logf("  内容: %s", article.Contents)
	}
}
