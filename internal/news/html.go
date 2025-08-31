package news

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"ncpd/internal/client"

	"github.com/PuerkitoBio/goquery"
)

// determineThumbnailURL 根据文章布局类型决定缩略图URL
func determineThumbnailURL(article *Article, channelThumbnailURL string) string {
	// 无布局信息，使用频道默认封面
	if article.ArticleTheme == nil || article.ArticleTheme.ArticleListLayoutType == nil {
		return channelThumbnailURL
	}

	layoutID := article.ArticleTheme.ArticleListLayoutType.ID
	if layoutID == 3 {
		// リスト型：强制使用频道默认封面
		return channelThumbnailURL
	}

	// 其他情况，优先使用文章封面，没有则用频道默认封面
	if article.ThumbnailURL != "" {
		fmt.Printf("使用文章封面: %s\n", article.ThumbnailURL)
		return article.ThumbnailURL
	}
	return channelThumbnailURL
}

// ProcessArticleWithOutputDir 处理文章并指定图片输出目录
func ProcessArticleWithOutputDir(article *Article, templateHTML, outputDir string, channelThumbnailURL string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(templateHTML))
	if err != nil {
		return "", err
	}

	// 下载图片并替换URL
	processedContents, err := downloadAndReplaceImages(article.Contents, outputDir)
	if err != nil {
		return "", fmt.Errorf("处理图片时出错: %w", err)
	}

	// 根据布局策略确定缩略图URL并下载
	thumbnailURL := determineThumbnailURL(article, channelThumbnailURL)
	if thumbnailURL != "" {
		thumbnailPath, err := downloadThumbnail(thumbnailURL, outputDir)
		if err != nil {
			fmt.Printf("下载缩略图失败: %v\n", err)
		} else {
			doc.Find(".Thumbnail img").SetAttr("src", thumbnailPath)
		}
	}

	// 处理发布时间格式，只保留日期部分
	publishDate := formatPublishDate(article.PublishAt)

	// 处理内容，替换特定的按钮标签
	processedContents = replaceButtonTags(processedContents)

	// 插入数据
	doc.Find(".Title h6").SetText(article.ArticelTitle)
	doc.Find(".PublishAt").SetHtml(`<span>` + publishDate + `</span>`)
	doc.Find(".Content").SetHtml(processedContents)

	// 获取最终HTML并解码实体
	finalHTML, _ := doc.Html()
	decodedHTML := html.UnescapeString(finalHTML)
	return decodedHTML, nil
}

// 下载图片并替换URL
func downloadAndReplaceImages(contents, customOutputDir string) (string, error) {
	// 先解码HTML实体
	decodedContents := html.UnescapeString(contents)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(decodedContents))
	if err != nil {
		return "", err
	}

	// 确定输出目录
	outputDir := "./out/NEWS"
	if customOutputDir != "" {
		outputDir = customOutputDir
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			// 下载图片并获取本地路径
			localPath, err := downloadImage(client, src, outputDir)
			if err != nil {
				fmt.Printf("下载图片失败 %s: %v\n", src, err)
				return
			}

			// 替换为本地路径
			s.SetAttr("src", localPath)
		}
	})

	htmlContent, _ := doc.Html()
	return htmlContent, nil
}

// 下载单个图片
func downloadImage(c *http.Client, imageURL, outputDir string) (string, error) {

	// 创建请求
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 添加Referer头
	req.Header.Set("Referer", fmt.Sprintf("https://%s/", client.CurrentPlatform.Domain))

	// 发送HTTP请求
	resp, err := c.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求图片失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode)
	}

	// 从URL中提取文件名
	fileName := extractFileName(imageURL)
	if fileName == "" {
		fileName = fmt.Sprintf("image_%d.jpg", time.Now().Unix())
	}

	// 构建本地文件路径
	localPath := filepath.Join(outputDir, fileName)

	// 创建文件
	file, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 写入文件
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	// 返回相对路径
	return "./" + fileName, nil
}

// 从URL中提取文件名
func extractFileName(imageURL string) string {
	// 移除查询参数
	if idx := strings.Index(imageURL, "?"); idx != -1 {
		imageURL = imageURL[:idx]
	}

	// 获取路径的最后一部分
	parts := strings.Split(imageURL, "/")
	if len(parts) > 0 {
		fileName := parts[len(parts)-1]
		if fileName != "" {
			// 解码URL编码的字符
			decodedFileName, err := url.QueryUnescape(fileName)
			if err == nil {
				fileName = decodedFileName
			}

			// 清理文件名，移除或替换不安全的字符
			fileName = cleanFileName(fileName)
			return fileName
		}
	}

	return ""
}

// 清理文件名，移除不安全的字符
func cleanFileName(fileName string) string {
	// 定义不允许的字符（适用于 Windows、macOS、Linux）
	invalidChars := regexp.MustCompile(`[<>:"|?*\x00-\x1f\x7f/\\]`)

	// 替换不允许的字符为下划线
	cleanName := invalidChars.ReplaceAllString(fileName, "_")

	// 移除开头和结尾的空格、点号
	cleanName = strings.Trim(cleanName, " .")

	// 如果清理后为空，使用默认名称
	if cleanName == "" {
		cleanName = fmt.Sprintf("image_%d", time.Now().Unix())
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

// replaceButtonTags 替换特定的按钮标签
func replaceButtonTags(contents string) string {
	// 替换 <p><button>ココマデ</button></p> 为 <br>
	re := regexp.MustCompile(`<p>\s*<button[^>]*>ココマデ</button>\s*</p>`)
	return re.ReplaceAllString(contents, "<br>")
}

// downloadThumbnail 下载缩略图
func downloadThumbnail(thumbnailURL, outputDir string) (string, error) {
	// 创建HTTP客户端
	c := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 创建请求
	req, err := http.NewRequest("GET", thumbnailURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 添加Referer头
	req.Header.Set("Referer", fmt.Sprintf("https://%s/", client.CurrentPlatform.Domain))

	// 发送HTTP请求
	resp, err := c.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求缩略图失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode)
	}

	// 从URL中提取文件名，专门处理缩略图
	fileName := extractThumbnailFileName(thumbnailURL)

	// 构建本地文件路径
	localPath := filepath.Join(outputDir, fileName)

	// 创建文件
	file, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("创建缩略图文件失败: %w", err)
	}
	defer file.Close()

	// 写入文件
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("写入缩略图文件失败: %w", err)
	}

	// 返回相对路径
	return "./" + fileName, nil
}

// extractThumbnailFileName 专门处理缩略图文件名
func extractThumbnailFileName(thumbnailURL string) string {
	// 统一命名为 thumbnail.jpg
	return "thumbnail.jpg"
}
