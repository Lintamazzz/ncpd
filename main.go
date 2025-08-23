package main

import (
	"encoding/json"
	"fmt"
	"io"
	"ncpd/internal/auth"
	"ncpd/internal/channel"
	"ncpd/internal/m3u8"
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

func main() {
	// 0. 用户输入关键词，搜索并选择要下载的频道
	fcSiteID, err := selectChannelDomain()
	if err != nil {
		fmt.Printf("❌ 选择频道失败: %v\n", err)
		return
	}

	// 1. 获取视频列表
	videoList, _ := video.GetVideoList(fcSiteID)
	fmt.Printf("\n=== 数据获取完成 ===\n")
	fmt.Printf("总共获取到 %d 个视频\n", len(videoList))

	// 2. 用户选择要下载的视频
	selectedVideos := selectVideos(videoList)
	if len(selectedVideos) == 0 {
		fmt.Println("\n❌ 未选择任何视频，程序退出")
		return
	}

	// 3. 确认下载
	if !confirmDownload(selectedVideos) {
		fmt.Println("\n❌ 用户取消下载，程序退出")
		return
	}

	// 4. 下载视频
	downloadVideos(selectedVideos)

	// 5. 保存视频详情信息
	saveVideoDetails(fcSiteID, selectedVideos)

	// 6. 下载视频封面
	downloadThumbnails(selectedVideos)

	// 7. 下载视频弹幕
	downloadDanmaku(selectedVideos)

	// 打印 refresh_token 用于后续的 token 刷新
	fmt.Printf("\n最近的 refresh_token: %s \n", auth.GetRefreshToken())
	fmt.Printf("请保存到 .env 文件中，用于后续的 token 刷新 \n")
}

func downloadVideos(selectedVideos []video.VideoDetails) {
	// 记录下载总耗时
	startTime := time.Now()
	// 记录成功、失败、跳过的视频数量
	var successCount, failCount, skipCount int
	// 记录失败的视频列表
	var failedVideos []string

	// 遍历选中的视频列表
	for i, video := range selectedVideos {
		// 确定保存路径和文件名
		saveDir, saveName := getSavePathAndName(video)

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

func saveVideoDetails(fcSiteID int, selectedVideos []video.VideoDetails) {
	// 记录成功和失败的视频数量
	var successCount, failCount int
	// 记录失败的视频列表
	var failedVideos []string

	for i, v := range selectedVideos {
		// 打印视频标题
		fmt.Printf("\n%d. %s\n", i+1, v.Title)

		// 确定保存路径和文件名
		saveDir, _ := getSavePathAndName(v)
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

func downloadThumbnails(selectedVideos []video.VideoDetails) {
	// 记录成功和失败的视频数量
	var successCount, failCount int
	// 记录失败的视频列表
	var failedVideos []string

	for i, video := range selectedVideos {
		// 打印视频标题
		fmt.Printf("\n%d. %s\n", i+1, video.Title)

		// 确定保存路径和文件名
		saveDir, _ := getSavePathAndName(video)
		saveName := "thumbnail"

		// 检查缩略图URL是否为空
		if video.ThumbnailURL == "" {
			fmt.Printf("❌ 缩略图URL为空\n")
			failCount++
			failedVideos = append(failedVideos, video.Title)
			continue
		}

		// 下载缩略图
		thumbnailFile := filepath.Join(saveDir, saveName+".jpg")
		if err := downloadImage(video.ThumbnailURL, thumbnailFile); err != nil {
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

func downloadDanmaku(selectedVideos []video.VideoDetails) {
	// todo
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
func getSavePathAndName(video video.VideoDetails) (string, string) {
	cleanTitle := sanitizeFilename(video.Title)

	if isLiveArchive(video) {
		// 生放送 archive：保存到 ./out/live/视频标题/
		saveDir := filepath.Join("./out", "live", cleanTitle)
		return saveDir, cleanTitle
	} else {
		// 普通视频：保存到 ./out/video/视频标题/
		saveDir := filepath.Join("./out", "video", cleanTitle)
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
				Title("请选择要下载的视频").
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

// confirmDownload 确认下载
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
				Title(fmt.Sprintf("确认下载以下 %d 个视频：\n", len(selectedVideos))).
				Description(videoListDesc.String()).
				Value(&confirmDownload),
		),
	)

	// 运行确认表单
	if err := form.Run(); err != nil {
		fmt.Printf("❌ 确认下载时出错: %v\n", err)
		return false
	}

	return confirmDownload
}
