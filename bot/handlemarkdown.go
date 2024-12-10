package main

import "strings"

func handleMarkdown(markdown string) string {
	// 分割 Markdown 文本
	lines := strings.Split(markdown, "\n")

	// 定义一个字符串来存储结果
	var result strings.Builder

	// 标记“内容类型”是否处理
	inContentType := false

	// 遍历每一行并进行处理
	for i := 0; i < len(lines); i++ {
		// 如果遇到 "关闭 Issue 前请先确认以下内容"，则停止处理
		if strings.Contains(lines[i], "关闭 Issue 前请先确认以下内容") {
			break
		}

		// 如果当前行是 "### 作者"
		if strings.TrimSpace(lines[i]) == "### 作者" {
			// 找到下一个非空行，将其视为“作者”的内容
			for j := i + 1; j < len(lines); j++ {
				if strings.TrimSpace(lines[j]) != "" {
					authorLine := "宣讲人：" + strings.TrimSpace(lines[j])
					result.WriteString(authorLine + "\n")
					i = j // 更新 i，跳过已处理的行
					break
				}
			}
			continue
		}

		// 如果当前行是 "### 内容类型"
		if strings.TrimSpace(lines[i]) == "### 内容类型" {
			inContentType = true
			continue
		}

		// 如果在 "内容类型" 部分，处理选项
		if inContentType {
			if strings.Contains(lines[i], "- [X] 技术类") {
				result.WriteString("内容类型：技术类\n")
				continue
			} else if strings.Contains(lines[i], "- [X] 其他") {
				result.WriteString("内容类型：其他\n")
				inContentType = false // 处理完内容类型，退出
				continue
			} else if strings.Contains(lines[i], "- [ ]") {
				// 忽略未选中的选项
				if !strings.Contains(lines[i+1], "- [X]") {
					inContentType = false
				}
				continue
			}
		}

		// 如果当前行是 "### 摘要/大纲"
		if strings.TrimSpace(lines[i]) == "### 摘要/大纲" {
			result.WriteString("摘要/大纲：\n")
			continue
		}

		// 如果当前行是 "### 补充说明"
		if strings.TrimSpace(lines[i]) == "### 补充说明" {
			// 找到下一个非空行，将其视为“作者”的内容
			for j := i + 1; j < len(lines); j++ {
				if strings.TrimSpace(lines[j]) != "" {
					if strings.TrimSpace(lines[j]) == "_No response_" {
						result.WriteString("补充说明：无\n")
						i = j // 更新 i，跳过已处理的行
						break
					} else {
						result.WriteString("补充说明：\n")
						break
					}
				}
			}
			continue
		}

		// 清理当前行
		cleaned := cleanMarkdown(lines[i])

		// 如果清理后的行不为空，添加到结果中
		if cleaned != "" {
			result.WriteString(cleaned + "\n")
		}
	}

	// 将最终结果返回为字符串
	return result.String()
}
