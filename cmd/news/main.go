package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"ncpd/internal/auth"
	"ncpd/internal/news"
)

func main() {
	// 1. 获取 token
	token, err := auth.GetToken()
	if err != nil {
		log.Fatal("获取Token失败:", err)
	}

	// 2. 指定频道ID
	fcSiteID := 387

	// 3. 获取文章列表
	fmt.Println("🔍 正在获取文章列表...")
	articles, err := news.GetArticleList(fcSiteID)
	if err != nil {
		log.Fatal("获取文章列表失败:", err)
	}

	fmt.Printf("✅ 获取到 %d 篇文章\n", len(articles))

	// 4. 读取HTML模板
	templateHTML, err := os.ReadFile("template.html")
	if err != nil {
		log.Fatal("读取模板文件失败:", err)
	}

	// 5. 遍历每篇文章，获取详细信息并生成HTML
	var successCount, failCount int
	var failedArticles []string

	for i, articleSummary := range articles {
		fmt.Printf("\n%d. 处理文章: %s\n", i+1, articleSummary.ArticelTitle)

		// 获取文章的详细信息
		article, err := news.GetArticle(fcSiteID, articleSummary.ArticleCode, token)
		if err != nil {
			fmt.Printf("❌ 获取文章详情失败: %v\n", err)
			failCount++
			failedArticles = append(failedArticles, articleSummary.ArticelTitle)
			continue
		}

		// 检查文章内容是否为空
		if article.Contents == "" {
			fmt.Printf("⚠️  文章内容为空，跳过处理\n")
			failCount++
			failedArticles = append(failedArticles, articleSummary.ArticelTitle)
			continue
		}

		// 生成HTML文件
		if err := generateArticleHTML(article, string(templateHTML)); err != nil {
			fmt.Printf("❌ 生成HTML失败: %v\n", err)
			failCount++
			failedArticles = append(failedArticles, articleSummary.ArticelTitle)
			continue
		}

		fmt.Printf("✅ 文章处理完成\n")
		successCount++
	}

	// 6. 打印统计信息
	fmt.Printf("\n" + strings.Repeat("=", 50) + "\n")
	fmt.Printf("处理完成！\n")
	fmt.Printf("成功处理: %d 篇文章\n", successCount)
	fmt.Printf("处理失败: %d 篇文章\n", failCount)
	fmt.Printf("总计: %d 篇文章\n", len(articles))

	if failCount > 0 {
		fmt.Printf("\n失败的文章列表:\n")
		for i, title := range failedArticles {
			fmt.Printf("  %d. %s\n", i+1, title)
		}
	}
	fmt.Printf(strings.Repeat("=", 50) + "\n")
}

// generateArticleHTML 为单篇文章生成HTML文件
func generateArticleHTML(article *news.Article, templateHTML string) error {
	// 清理文章标题作为文件夹名
	cleanTitle := sanitizeFilename(article.ArticelTitle)

	// 格式化发布时间，只保留日期部分（用于目录名，保持 2024-12-18 格式）
	var publishDate string
	if len(article.PublishAt) >= 10 {
		publishDate = article.PublishAt[:10] // 保持 2024-12-18 格式
	} else {
		publishDate = article.PublishAt
	}

	// 构建目录名：格式为 "[2024-12-18] article_title"
	dirName := fmt.Sprintf("[%s] %s", publishDate, cleanTitle)

	// 创建输出目录
	outputDir := filepath.Join("out", "NEWS", dirName)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 生成HTML内容，图片保存到文章目录
	html, err := news.ProcessArticleWithOutputDir(article, templateHTML, outputDir)
	if err != nil {
		return fmt.Errorf("处理文章失败: %w", err)
	}

	// 保存HTML文件
	htmlFilePath := filepath.Join(outputDir, cleanTitle+".html")
	if err := os.WriteFile(htmlFilePath, []byte(html), 0644); err != nil {
		return fmt.Errorf("保存HTML文件失败: %w", err)
	}

	fmt.Printf("   📄 HTML文件: %s\n", htmlFilePath)
	return nil
}

// sanitizeFilename 清理文件名，移除或替换特殊字符，使其适用于所有平台
func sanitizeFilename(filename string) string {
	// 定义不允许的字符（适用于 Windows、macOS、Linux）
	invalidChars := regexp.MustCompile(`[<>:"|?*\x00-\x1f\x7f/\\]`)

	// 替换不允许的字符为下划线
	cleanName := invalidChars.ReplaceAllString(filename, "_")

	// 移除开头和结尾的空格、点号
	cleanName = strings.Trim(cleanName, " .")

	// 如果清理后为空，使用默认名称
	if cleanName == "" {
		cleanName = fmt.Sprintf("article_%d", time.Now().Unix())
	}

	// 限制长度（避免路径过长）
	if len(cleanName) > 200 {
		cleanName = cleanName[:200]
	}

	return cleanName
}

// formatPublishDate 格式化发布时间，只保留日期部分，并将格式改为 2024/12/18
func formatPublishDate(publishAt string) string {
	// 如果包含时间，只取日期部分
	if len(publishAt) >= 10 {
		datePart := publishAt[:10]
		// 将 2024-12-18 格式改为 2024/12/18
		return strings.ReplaceAll(datePart, "-", "/")
	}
	return publishAt
}
