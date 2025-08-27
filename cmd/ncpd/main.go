package main

import (
	"encoding/json"
	"fmt"
	"io"
	"ncpd/internal/auth"
	"ncpd/internal/channel"
	"ncpd/internal/m3u8"
	"ncpd/internal/news"
	"ncpd/internal/video"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
)

// DownloadOptions 定义用户选择的下载选项
type DownloadOptions struct {
	Video        bool
	VideoDetails bool
	Thumbnail    bool
	Danmaku      bool
	News         bool
}

// HasAnySelection 检查是否有任何选择
func (d *DownloadOptions) HasAnySelection() bool {
	return d.Video || d.VideoDetails || d.Thumbnail || d.Danmaku || d.News
}

func main() {
	// 0. 用户输入关键词，搜索并选择要下载的频道
	fcSiteID, err := selectChannelDomain()
	if err != nil {
		fmt.Printf("❌ 选择频道失败: %v\n", err)
		return
	}

	// 获取频道信息
	fmt.Println("🔍 正在获取频道信息...")
	channelInfo, err := channel.GetFanclubSiteInfo(fcSiteID)
	if err != nil {
		fmt.Printf("❌ 获取频道信息失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 频道信息获取成功: %s\n", channelInfo.FanclubSiteName)

	// 创建基础保存目录
	channelName := sanitizeFilename(channelInfo.FanclubSiteName)
	baseSaveDir := filepath.Join("./out", channelName)
	fmt.Printf("📁 保存目录: %s\n", baseSaveDir)

	// 1. 首先询问用户要下载什么类型的内容
	downloadOptions := selectDownloadOptions()
	if !downloadOptions.HasAnySelection() {
		fmt.Println("\n❌ 未选择任何下载内容，程序退出")
		return
	}

	// 2. 根据选择的内容类型执行相应的操作

	// 获取频道默认封面地址
	defaultThumbnailURL := ""
	if channelInfo != nil {
		defaultThumbnailURL = channelInfo.ThumbnailImageURL
	}

	// 如果选择了新闻，先下载新闻
	if downloadOptions.News {
		if !confirmNewsDownload() {
			fmt.Println("\n❌ 用户取消下载新闻，程序退出")
			return
		}
		downloadNews(baseSaveDir, fcSiteID, defaultThumbnailURL)
	}

	// 如果选择了视频相关的内容，需要获取视频列表
	if downloadOptions.Video || downloadOptions.VideoDetails || downloadOptions.Thumbnail || downloadOptions.Danmaku {
		videoList, _ := video.GetVideoList(fcSiteID)
		fmt.Printf("\n=== 数据获取完成 ===\n")
		fmt.Printf("总共获取到 %d 个视频\n", len(videoList))

		// 用户选择要下载的视频
		selectedVideos := selectVideos(videoList)
		if len(selectedVideos) == 0 {
			fmt.Println("\n❌ 未选择任何视频，程序退出")
			return
		}

		// 确认下载
		if !confirmDownload(selectedVideos) {
			fmt.Println("\n❌ 用户取消下载，程序退出")
			return
		}

		// 根据选择执行相应的下载任务
		if downloadOptions.Video {
			downloadVideos(baseSaveDir, selectedVideos)
		}

		if downloadOptions.VideoDetails {
			saveVideoDetails(baseSaveDir, fcSiteID, selectedVideos)
		}

		if downloadOptions.Thumbnail {
			downloadThumbnails(baseSaveDir, selectedVideos, defaultThumbnailURL)
		}

		if downloadOptions.Danmaku {
			downloadDanmaku(baseSaveDir, fcSiteID, selectedVideos)
		}
	}

	// 打印 refresh_token 用于后续的 token 刷新
	fmt.Printf("\n最近的 refresh_token: %s \n", auth.GetRefreshToken())
	fmt.Printf("请保存到 .env 文件中，用于后续的 token 刷新 \n")
}

func downloadVideos(baseSaveDir string, selectedVideos []video.VideoDetails) {
	// 记录下载总耗时
	startTime := time.Now()
	// 记录成功、失败、跳过的视频数量
	var successCount, failCount, skipCount int
	// 记录失败的视频列表
	var failedVideos []string

	// 遍历选中的视频列表
	for i, video := range selectedVideos {
		// 确定保存路径和文件名
		saveDir, saveName := getSavePathAndName(video, baseSaveDir)

		// 检查视频文件是否已经存在，如果存在则跳过下载
		expectedFile := filepath.Join(saveDir, saveName+".ts")
		if _, err := os.Stat(expectedFile); err == nil {
			fmt.Printf("\n%d. %s\n", i+1, video.Title)
			fmt.Printf("   文件已存在，跳过下载: %s\n", expectedFile)
			skipCount++
			continue
		}

		// 记录单个文件下载开始时间
		fileStartTime := time.Now()

		// 获取 token（有效期为 5 分钟）
		token, err := auth.GetToken()
		if err != nil {
			fmt.Printf("\n%d. %s\n", i+1, video.Title)
			fmt.Printf("   ❌ 获取 Token 失败: %v\n", err)
			failCount++
			failedVideos = append(failedVideos, video.Title)
			continue
		}

		sessionID, err := auth.GetSessionID(video.ContentCode, token)
		if err != nil {
			fmt.Printf("\n%d. %s\n", i+1, video.Title)
			fmt.Printf("   ❌ 获取 sessionID 失败: %v\n", err)
			failCount++
			failedVideos = append(failedVideos, video.Title)
			continue
		}
		index, err := m3u8.GetIndex(sessionID)
		if err != nil {
			fmt.Printf("\n%d. %s\n", i+1, video.Title)
			fmt.Printf("   ❌ 获取 index.m3u8 失败: %v\n", err)
			failCount++
			failedVideos = append(failedVideos, video.Title)
			continue
		}
		streamInfo := m3u8.ParseIndexM3U8(index)
		bestQuality := m3u8.GetBestQuality(streamInfo)

		fmt.Printf("\n%d. %s\n", i+1, video.Title)
		fmt.Printf("   视频代码: %s\n", video.ContentCode)
		fmt.Printf("   最高画质: %s %s\n", bestQuality.Resolution, bestQuality.FrameRate)
		fmt.Printf("   下载地址: %s\n", bestQuality.URL)
		fmt.Printf("   开始执行下载...\n\n")

		// 执行下载并检查结果
		if err = downloadVideo(bestQuality.URL, saveDir, saveName); err == nil {
			// 计算单个文件下载耗时
			fileDuration := time.Since(fileStartTime)
			fmt.Printf("\n   ✅ 下载成功，耗时: %s\n", formatDuration(fileDuration))
			successCount++
		} else {
			fmt.Printf("\n   ❌ 下载失败: %v\n", err)
			failCount++
			failedVideos = append(failedVideos, video.Title)
			continue
		}
	}

	// 计算总耗时
	totalDuration := time.Since(startTime)

	// 打印最终统计信息
	fmt.Printf("\n" + strings.Repeat("=", 50) + "\n")
	fmt.Printf("下载完成！总耗时: %s\n", formatDuration(totalDuration))
	fmt.Printf("成功下载: %d 个文件\n", successCount)
	fmt.Printf("下载失败: %d 个文件\n", failCount)
	fmt.Printf("跳过下载: %d 个文件\n", skipCount)
	fmt.Printf("总计: %d 个文件\n", len(selectedVideos))

	if failCount > 0 {
		fmt.Printf("\n失败的文件列表:\n")
		for i, title := range failedVideos {
			fmt.Printf("  %d. %s\n", i+1, title)
		}
	}
	fmt.Printf(strings.Repeat("=", 50) + "\n")
}

func downloadVideo(url string, saveDir string, saveName string) error {
	// 确保保存目录存在
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	cmd := exec.Command("N_m3u8DL-RE", url,
		"-H", "User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36",
		"--save-dir", saveDir,
		"--save-name", saveName,
		"--binary-merge", // 防止 ts 分片过多导致合并时报错，开启后输出文件由 .mp4 变为 .ts
	)

	// 实时显示下载进度
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("命令执行失败: %w", err)
	}

	return nil
}

func saveVideoDetails(baseSaveDir string, fcSiteID int, selectedVideos []video.VideoDetails) {
	// 记录成功和失败的视频数量
	var successCount, failCount int
	// 记录失败的视频列表
	var failedVideos []string

	for i, v := range selectedVideos {
		// 打印视频标题
		fmt.Printf("\n%d. %s\n", i+1, v.Title)

		// 确定保存路径和文件名
		saveDir, _ := getSavePathAndName(v, baseSaveDir)
		saveName := "video_details"

		// 获取视频的详细信息
		videoDetails, err := video.GetVideoDetails(fcSiteID, v.ContentCode)
		if err != nil {
			fmt.Printf("❌ 获取视频详情失败: %v\n", err)
			failCount++
			failedVideos = append(failedVideos, v.Title)
			continue
		}

		// 保存视频的详细信息，如果文件存在的话直接覆盖
		videoJSON, err := json.MarshalIndent(videoDetails, "", "  ")
		if err != nil {
			fmt.Printf("❌ JSON 序列化失败: %v\n", err)
			failCount++
			failedVideos = append(failedVideos, v.Title)
			continue
		}
		// 确保保存目录存在
		if err := os.MkdirAll(saveDir, 0755); err != nil {
			fmt.Printf("❌ 创建目录失败: %v\n", err)
			failCount++
			failedVideos = append(failedVideos, v.Title)
			continue
		}
		// 保存视频详情文件
		videoFile := filepath.Join(saveDir, saveName+".json")
		if err := os.WriteFile(videoFile, videoJSON, 0644); err != nil {
			fmt.Printf("❌ 保存视频详情失败: %v\n", err)
			failCount++
			failedVideos = append(failedVideos, v.Title)
			continue
		}
		fmt.Printf("✅ 已保存视频详情: %s\n", videoFile)
		successCount++
	}

	// 打印最终统计信息
	fmt.Printf("\n" + strings.Repeat("=", 50) + "\n")
	fmt.Printf("成功保存: %d 个视频详情文件\n", successCount)
	fmt.Printf("保存失败: %d 个视频详情文件\n", failCount)
	fmt.Printf("总计: %d 个视频详情文件\n", len(selectedVideos))

	if failCount > 0 {
		fmt.Printf("\n失败的文件列表:\n")
		for i, title := range failedVideos {
			fmt.Printf("  %d. %s\n", i+1, title)
		}
	}
	fmt.Printf(strings.Repeat("=", 50) + "\n")
}

func downloadThumbnails(baseSaveDir string, selectedVideos []video.VideoDetails, defaultThumbnailURL string) {
	// 记录成功和失败的视频数量
	var successCount, failCount int
	// 记录失败的视频列表
	var failedVideos []string

	for i, video := range selectedVideos {
		// 打印视频标题
		fmt.Printf("\n%d. %s\n", i+1, video.Title)

		// 确定保存路径和文件名
		saveDir, _ := getSavePathAndName(video, baseSaveDir)
		saveName := "thumbnail"

		// 确定要下载的缩略图URL
		thumbnailURL := video.ThumbnailURL
		if thumbnailURL == "" {
			if defaultThumbnailURL != "" {
				thumbnailURL = defaultThumbnailURL
				fmt.Printf("   使用频道默认封面: %s\n", thumbnailURL)
			} else {
				fmt.Printf("❌ 缩略图URL为空且无频道默认封面\n")
				failCount++
				failedVideos = append(failedVideos, video.Title)
				continue
			}
		}

		// 下载缩略图
		thumbnailFile := filepath.Join(saveDir, saveName+".jpg")
		if err := downloadImage(thumbnailURL, thumbnailFile); err != nil {
			fmt.Printf("❌ 下载缩略图失败: %v\n", err)
			failCount++
			failedVideos = append(failedVideos, video.Title)
			continue
		}
		fmt.Printf("✅ 已保存缩略图: %s\n", thumbnailFile)
		successCount++
	}

	// 打印最终统计信息
	fmt.Printf("\n" + strings.Repeat("=", 50) + "\n")
	fmt.Printf("成功下载: %d 个缩略图\n", successCount)
	fmt.Printf("下载失败: %d 个缩略图\n", failCount)
	fmt.Printf("总计: %d 个缩略图\n", len(selectedVideos))

	if failCount > 0 {
		fmt.Printf("\n失败的缩略图列表:\n")
		for i, title := range failedVideos {
			fmt.Printf("  %d. %s\n", i+1, title)
		}
	}
	fmt.Printf(strings.Repeat("=", 50) + "\n")
}

func downloadImage(url string, filePath string) error {
	// 确保保存目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 发送HTTP请求下载图片
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode)
	}

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 将响应内容写入文件
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

func downloadDanmaku(baseSaveDir string, fcSiteID int, selectedVideos []video.VideoDetails) {
	// 记录成功和失败的视频数量
	var successCount, failCount int
	// 记录失败的视频列表
	var failedVideos []string

	for i, v := range selectedVideos {
		// 打印视频标题
		fmt.Printf("\n%d. %s\n", i+1, v.Title)

		// 确定保存路径和文件名
		saveDir, _ := getSavePathAndName(v, baseSaveDir)
		saveName := "danmaku"

		details, err := video.GetVideoDetails(fcSiteID, v.ContentCode)
		if err != nil {
			fmt.Printf("❌ 获取视频详情失败: %v\n", err)
			failCount++
			failedVideos = append(failedVideos, v.Title)
			continue
		}

		// 检查是否有评论设置
		if details.VideoCommentSetting == nil || details.VideoCommentSetting.CommentGroupID == "" {
			fmt.Printf("❌ 视频没有评论设置或评论组ID为空\n")
			failCount++
			failedVideos = append(failedVideos, v.Title)
			continue
		}

		// 获取评论用户token
		commentsUserToken, err := video.GetCommentsUserToken(v.ContentCode)
		if err != nil {
			fmt.Printf("❌ 获取评论用户token失败: %v\n", err)
			failCount++
			failedVideos = append(failedVideos, v.Title)
			continue
		}

		// 获取所有弹幕
		fmt.Printf("  获取弹幕中...\n")
		allComments, err := video.GetAllComments(commentsUserToken, details.VideoCommentSetting.CommentGroupID)
		if err != nil {
			fmt.Printf("❌ 获取弹幕失败: %v\n", err)
			failCount++
			failedVideos = append(failedVideos, v.Title)
			continue
		}

		// 保存弹幕为JSON文件
		commentsJSON, err := json.MarshalIndent(allComments, "", "  ")
		if err != nil {
			fmt.Printf("❌ JSON序列化失败: %v\n", err)
			failCount++
			failedVideos = append(failedVideos, v.Title)
			continue
		}

		// 确保保存目录存在
		if err := os.MkdirAll(saveDir, 0755); err != nil {
			fmt.Printf("❌ 创建目录失败: %v\n", err)
			failCount++
			failedVideos = append(failedVideos, v.Title)
			continue
		}

		// 保存弹幕文件
		danmakuFile := filepath.Join(saveDir, saveName+".json")
		if err := os.WriteFile(danmakuFile, commentsJSON, 0644); err != nil {
			fmt.Printf("❌ 保存弹幕失败: %v\n", err)
			failCount++
			failedVideos = append(failedVideos, v.Title)
			continue
		}

		fmt.Printf("✅ 已保存弹幕: %s (共 %d 条)\n", danmakuFile, len(allComments))
		successCount++
	}

	// 打印最终统计信息
	fmt.Printf("\n" + strings.Repeat("=", 50) + "\n")
	fmt.Printf("成功下载: %d 个弹幕文件\n", successCount)
	fmt.Printf("下载失败: %d 个弹幕文件\n", failCount)
	fmt.Printf("总计: %d 个弹幕文件\n", len(selectedVideos))

	if failCount > 0 {
		fmt.Printf("\n失败的弹幕文件列表:\n")
		for i, title := range failedVideos {
			fmt.Printf("  %d. %s\n", i+1, title)
		}
	}
	fmt.Printf(strings.Repeat("=", 50) + "\n")
}

func downloadNews(baseSaveDir string, fcSiteID int, defaultThumbnailURL string) {
	fmt.Printf("\n=== 开始下载频道新闻 ===\n")

	// 获取 token
	token, err := auth.GetToken()
	if err != nil {
		fmt.Printf("❌ 获取 Token 失败: %v\n", err)
		return
	}

	// 获取文章列表
	fmt.Println("🔍 正在获取文章列表...")
	articles, err := news.GetArticleList(fcSiteID)
	if err != nil {
		fmt.Printf("❌ 获取文章列表失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 获取到 %d 篇文章\n", len(articles))

	// 读取HTML模板
	templateHTML, err := os.ReadFile("template.html")
	if err != nil {
		fmt.Printf("❌ 读取模板文件失败: %v\n", err)
		return
	}

	// 处理每篇文章
	var successCount, failCount int
	var failedArticles []string

	for i, articleSummary := range articles {
		fmt.Printf("\n%d. 处理文章: %s\n", i+1, articleSummary.ArticelTitle)

		// 获取文章详细信息
		article, err := news.GetArticle(fcSiteID, articleSummary.ArticleCode, token)
		if err != nil {
			fmt.Printf("❌ 获取文章详情失败: %v\n", err)
			failCount++
			failedArticles = append(failedArticles, articleSummary.ArticelTitle)
			continue
		}

		// 检查文章内容是否为空
		if article.Contents == "" {
			fmt.Printf("⚠️  会员限定内容，跳过处理\n")
			failCount++
			failedArticles = append(failedArticles, articleSummary.ArticelTitle)
			continue
		}

		// 生成HTML文件
		if err := generateArticleHTML(article, string(templateHTML), baseSaveDir, defaultThumbnailURL); err != nil {
			fmt.Printf("❌ 生成HTML失败: %v\n", err)
			failCount++
			failedArticles = append(failedArticles, articleSummary.ArticelTitle)
			continue
		}

		fmt.Printf("✅ 文章处理完成\n")
		successCount++
	}

	// 打印统计信息
	fmt.Printf("\n" + strings.Repeat("=", 50) + "\n")
	fmt.Printf("新闻下载完成！\n")
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
func generateArticleHTML(article *news.Article, templateHTML string, baseSaveDir string, defaultThumbnailURL string) error {
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
	outputDir := filepath.Join(baseSaveDir, "NEWS", dirName)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 生成HTML内容，图片保存到文章目录
	html, err := news.ProcessArticleWithOutputDir(article, templateHTML, outputDir, defaultThumbnailURL)
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

// isLiveArchive 判断视频是否为生放送アーカイブ
func isLiveArchive(video video.VideoDetails) bool {
	// 检查 ActiveVideoFilename.VideoFilenameType.Value 是否为 "archived"
	if video.ActiveVideoFilename != nil &&
		video.ActiveVideoFilename.VideoFilenameType != nil &&
		video.ActiveVideoFilename.VideoFilenameType.Value == "archived" {
		return true
	}

	// 检查 LiveStartedAt 是否不为 nil
	if video.LiveStartedAt != nil {
		return true
	}

	return false
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
		cleanName = fmt.Sprintf("untitled_%d", time.Now().Unix())
	}

	// 限制长度（避免路径过长）
	if len(cleanName) > 200 {
		cleanName = cleanName[:200]
	}

	return cleanName
}

// getSavePathAndName 根据视频类型确定保存路径和文件名
func getSavePathAndName(video video.VideoDetails, baseSaveDir string) (string, string) {
	cleanTitle := sanitizeFilename(video.Title)

	if isLiveArchive(video) {
		// 生放送 archive：保存到 baseSaveDir/生放送/视频标题/
		saveDir := filepath.Join(baseSaveDir, "生放送", cleanTitle)
		return saveDir, cleanTitle
	} else {
		// 普通视频：保存到 baseSaveDir/動画/视频标题/
		saveDir := filepath.Join(baseSaveDir, "動画", cleanTitle)
		return saveDir, cleanTitle
	}
}

// formatDuration 格式化时间显示，便于阅读
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f秒", d.Seconds())
	} else if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%d分%d秒", minutes, seconds)
	} else {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%d小时%d分%d秒", hours, minutes, seconds)
	}
}

// selectChannelDomain 让用户选择频道并返回对应的ID
func selectChannelDomain() (int, error) {
	for {
		// 第一步，让用户输入搜索关键字
		var searchKeyword string
		searchForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("请输入频道域名或关键字").
					Placeholder(`"https://nicochannel.jp/abcd" or "abc"`).
					Value(&searchKeyword).
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("搜索关键字不能为空")
						}
						return nil
					}),
			),
		)

		// 运行搜索表单
		if err := searchForm.Run(); err != nil {
			fmt.Printf("❌ 输入搜索关键字时出错: %v\n", err)
			return -1, err
		}

		// 获取完整的频道列表
		fmt.Println("🔍 正在获取频道列表...")
		channels, err := channel.GetChannelList()
		if err != nil {
			fmt.Printf("❌ 获取频道列表失败: %v\n", err)
			return -1, err
		}

		// 根据关键字过滤频道
		var matchedChannels []channel.ContentProvider
		for _, ch := range channels {
			// 匹配域名（完整或部分）
			if strings.Contains(strings.ToLower(ch.Domain), strings.ToLower(searchKeyword)) {
				matchedChannels = append(matchedChannels, ch)
			}
		}

		if len(matchedChannels) == 0 {
			fmt.Printf("❌ 未找到匹配 '%s' 的频道\n", searchKeyword)

			// 询问是否重新输入
			var retry bool
			retryForm := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title("是否重新输入？").
						Value(&retry),
				),
			)

			if err := retryForm.Run(); err != nil {
				return -1, err
			}

			if retry {
				continue // 重新开始循环
			} else {
				return -1, fmt.Errorf("用户取消选择")
			}
		}

		// 如果只有一个匹配结果，也需要确认
		if len(matchedChannels) == 1 {
			selectedChannel := matchedChannels[0]

			// 确认选择的频道
			var confirmSelection bool
			confirmForm := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("找到频道: %s (ID: %d)\n是否确认选择该频道？", selectedChannel.Domain, selectedChannel.FanclubSite.ID)).
						Value(&confirmSelection),
				),
			)

			// 运行确认表单
			if err := confirmForm.Run(); err != nil {
				fmt.Printf("❌ 确认选择时出错: %v\n", err)
				return -1, err
			}

			if confirmSelection {
				fmt.Printf("✅ 已确认选择频道: %s (ID: %d)\n", selectedChannel.Domain, selectedChannel.FanclubSite.ID)
				return selectedChannel.FanclubSite.ID, nil
			} else {
				fmt.Println("❌ 用户取消选择，程序退出")
				return -1, fmt.Errorf("用户取消选择")
			}
		}

		// 创建选项列表，添加"重新输入"选项
		var options []huh.Option[int]
		for _, ch := range matchedChannels {
			options = append(options, huh.Option[int]{
				Key:   fmt.Sprintf("%s (ID: %d)", ch.Domain, ch.FanclubSite.ID),
				Value: ch.FanclubSite.ID,
			})
		}
		// 添加"重新输入"选项，使用特殊值 -1
		options = append(options, huh.Option[int]{
			Key:   "🔄 重新输入搜索关键字",
			Value: -1,
		})

		// 第二步：让用户从匹配的频道中选择
		var selectedChannelID int
		selectForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[int]().
					Title(fmt.Sprintf("找到 %d 个匹配的频道，请选择:", len(matchedChannels))).
					Options(options...).
					Value(&selectedChannelID),
			),
		)

		// 运行选择表单
		if err := selectForm.Run(); err != nil {
			fmt.Printf("❌ 选择频道时出错: %v\n", err)
			return -1, err
		}

		// 检查是否选择了"重新输入"
		if selectedChannelID == -1 {
			continue // 重新开始循环
		}

		// 找到选中的频道信息
		var selectedChannel channel.ContentProvider
		for _, ch := range matchedChannels {
			if ch.FanclubSite.ID == selectedChannelID {
				selectedChannel = ch
				break
			}
		}

		// 第三步：确认选择的频道
		var confirmSelection bool
		confirmForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(fmt.Sprintf("是否选择: %s (ID: %d) ?", selectedChannel.Domain, selectedChannelID)).
					Value(&confirmSelection),
			),
		)

		// 运行确认表单
		if err := confirmForm.Run(); err != nil {
			fmt.Printf("❌ 确认选择时出错: %v\n", err)
			return -1, err
		}

		if confirmSelection {
			fmt.Printf("✅ 已确认选择频道: %s (ID: %d)\n", selectedChannel.Domain, selectedChannelID)
			return selectedChannelID, nil
		} else {
			fmt.Println("❌ 用户取消选择，程序退出")
			return -1, fmt.Errorf("用户取消选择")
		}
	}
}

// selectVideos 让用户选择要下载的视频
func selectVideos(videoList []video.VideoDetails) []video.VideoDetails {
	// 创建选项列表
	var options []huh.Option[int]
	for i, video := range videoList {
		options = append(options, huh.Option[int]{
			Key:   video.Title,
			Value: i,
		})
	}

	// 创建多选表单
	var selectedIndices []int
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[int]().
				Title("请选择视频").
				Options(options...).
				Value(&selectedIndices),
		),
	)

	// 运行表单
	if err := form.Run(); err != nil {
		fmt.Printf("❌ 选择视频时出错: %v\n", err)
		return nil
	}

	// 根据选择的索引获取视频
	var selectedVideos []video.VideoDetails
	for _, index := range selectedIndices {
		if index >= 0 && index < len(videoList) {
			selectedVideos = append(selectedVideos, videoList[index])
		}
	}

	return selectedVideos
}

// selectDownloadOptions 让用户选择要下载的内容类型
func selectDownloadOptions() *DownloadOptions {
	options := &DownloadOptions{}

	// 创建选项列表
	var selectedOptions []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("请选择要下载的内容类型").
				Options(
					huh.Option[string]{Key: "视频", Value: "视频"},
					huh.Option[string]{Key: "视频封面", Value: "视频封面"},
					huh.Option[string]{Key: "视频弹幕", Value: "视频弹幕"},
					huh.Option[string]{Key: "视频详细信息", Value: "视频详细信息"},
					huh.Option[string]{Key: "频道新闻", Value: "频道新闻"},
				).
				Value(&selectedOptions),
		),
	)

	// 运行表单
	if err := form.Run(); err != nil {
		fmt.Printf("❌ 选择下载内容时出错: %v\n", err)
		return &DownloadOptions{}
	}

	// 根据选择设置结构体字段
	for _, option := range selectedOptions {
		switch option {
		case "视频":
			options.Video = true
		case "视频详细信息":
			options.VideoDetails = true
		case "视频封面":
			options.Thumbnail = true
		case "视频弹幕":
			options.Danmaku = true
		case "频道新闻":
			options.News = true
		}
	}

	return options
}

// confirmDownload 确认下载视频
func confirmDownload(selectedVideos []video.VideoDetails) bool {
	// 构建视频列表描述
	var videoListDesc strings.Builder
	for i, video := range selectedVideos {
		videoListDesc.WriteString(fmt.Sprintf("  %d. %s\n", i+1, video.Title))
	}

	var confirmDownload bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("确认选择以下 %d 个视频：\n", len(selectedVideos))).
				Description(videoListDesc.String()).
				Value(&confirmDownload),
		),
	)

	// 运行确认表单
	if err := form.Run(); err != nil {
		fmt.Printf("❌ 确认选择时出错: %v\n", err)
		return false
	}

	return confirmDownload
}

// confirmNewsDownload 确认下载新闻
func confirmNewsDownload() bool {
	var confirmDownload bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("确认下载频道所有新闻？").
				Value(&confirmDownload),
		),
	)

	// 运行确认表单
	if err := form.Run(); err != nil {
		fmt.Printf("❌ 确认下载新闻时出错: %v\n", err)
		return false
	}

	return confirmDownload
}
