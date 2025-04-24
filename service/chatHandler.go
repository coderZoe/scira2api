package service

import (
	"bufio"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"scira2api/config"
	"scira2api/log"
	"scira2api/models"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
)

const BASE_URL = "https://mcp.scira.ai"

type ChatHandler struct {
	Config *config.Config
	Client *req.Client
	index  int64
}

func NewChatHandler(config *config.Config) *ChatHandler {
	client := req.C().ImpersonateChrome().SetTimeout(time.Minute * 5).SetBaseURL(BASE_URL)
	if config.HttpProxy != "" {
		client.SetProxyURL(config.HttpProxy)
	}
	client.SetCommonHeader("Content-Type", "application/json")
	client.SetCommonHeader("Accept", "*/*")
	client.SetCommonHeader("Origin", BASE_URL)
	return &ChatHandler{
		Config: config,
		Client: client,
		index:  int64(rand.Intn(len(config.UserIds))),
	}
}

func (h *ChatHandler) ModelGetHandler(c *gin.Context) {
	data := make([]models.OpenAIModelResponse, len(h.Config.Models))
	for _, model := range h.Config.Models {
		model := models.OpenAIModelResponse{
			ID:      model,
			Created: time.Now().Unix(),
			Object:  "model",
		}
		data = append(data, model)
	}

	c.JSON(200, gin.H{
		"object": "list",
		"data":   data,
	})
}

// 处理请求
func (h *ChatHandler) ChatCompletionsHandler(c *gin.Context) {
	var request models.OpenAIChatCompletionsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error("bind json error: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.chatParamCheck(request)
	if err != nil {
		log.Error("chat param check error: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, chatId, userId, err := h.doChatRequest(request)
	if err != nil {
		log.Error("retry %d times request still error: %s", h.Config.Retry, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if request.Stream {
		h.handleStreamResponse(c, resp, request.Model)
	} else {
		h.handleRegularResponse(c, resp, request.Model)
	}
	if h.Config.ChatDelete {
		go h.deleteChat(chatId, userId)
	}
}

func (h *ChatHandler) chatParamCheck(request models.OpenAIChatCompletionsRequest) error {
	if request.Model == "" {
		return errors.New("model is required")
	}
	if !slices.Contains(h.Config.Models, request.Model) {
		return errors.New("model is not supported")
	}
	if len(request.Messages) == 0 {
		return errors.New("messages is required")
	}
	return nil
}

func (h *ChatHandler) getUserId() string {
	userIdsLength := int64(len(h.Config.UserIds))
	newIndex := atomic.AddInt64(&h.index, 1)
	return h.Config.UserIds[newIndex%userIdsLength]
}

func (h *ChatHandler) getChatId() string {
	// 生成15字节的随机数据（base64编码后约为20个字符）
	randomBytes := make([]byte, 15)
	crand.Read(randomBytes)
	// 使用base64编码，得到URL安全的字符串
	encoded := base64.RawURLEncoding.EncodeToString(randomBytes)

	// 确保长度为21（包括连字符）
	if len(encoded) < 20 {
		encoded = encoded + strings.Repeat("A", 20-len(encoded))
	} else if len(encoded) > 20 {
		encoded = encoded[:20]
	}

	// 在第11个位置后插入连字符
	return encoded[:11] + "-" + encoded[11:]
}

func (h *ChatHandler) doChatRequest(request models.OpenAIChatCompletionsRequest) (*req.Response, string, string, error) {
	var resp *req.Response
	var err error
	var chatId string
	var userId string
	for range h.Config.Retry {
		chatId = h.getChatId()
		userId = h.getUserId()
		log.Info("request use userId: %s, generate chatId: %s", userId, chatId)
		sciraRequest := request.ToSciraChatCompletionsRequest(request.Model, chatId, userId)
		resp, err = h.Client.R().SetHeader("Referer", BASE_URL+"/chat/"+chatId).SetBody(sciraRequest).Post(BASE_URL + "/api/chat")
		if err == nil {
			break
		}
		log.Error("userId: %s, chatId: %s, request error: %s", userId, chatId, err)
	}
	return resp, chatId, userId, err
}

func (h *ChatHandler) deleteChat(chatId string, userId string) {
	resp, err := h.Client.R().SetHeader("X-User-Id", userId).SetHeader("Referer", BASE_URL+"/chat/"+chatId).Delete(BASE_URL + "/api/chats/" + chatId)
	if err != nil {
		log.Error("userId: %s, chatId: %s, delete chat error: %s", userId, chatId, err)
	}
	if resp.StatusCode != 200 {
		log.Error("userId: %s, chatId: %s, delete chat status code: %d", userId, chatId, resp.StatusCode)
	}
}

// 处理流式响应
func (h *ChatHandler) handleStreamResponse(c *gin.Context, resp *req.Response, model string) {
	// 设置响应头
	defer resp.Body.Close()
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Status(http.StatusOK)

	id := fmt.Sprintf("chatcmpl-%s%s", time.Now().Format("20060102150405"), randString(10))
	createdTime := time.Now().Unix()

	// 发送流式响应头部
	oaiResponse := models.NewOaiStreamResponse(id, createdTime, model, nil)
	// 创建扫描器来处理响应流
	reader := bufio.NewScanner(resp.Body)
	clientDone := c.Request.Context().Done()
	for reader.Scan() {
		select {
		case <-clientDone:

			return
		default:
			//do nothing
		}
		line := reader.Text()
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "g:") {
			// 推理内容
			content := processContent(line[2:])
			oaiResponse.Choices = models.NewChoice("", content, "")
			h.writeStreamResponse(c, oaiResponse)
		} else if strings.HasPrefix(line, "0:") {
			// 常规内容
			content := processContent(line[2:])
			oaiResponse.Choices = models.NewChoice(content, "", "")
			h.writeStreamResponse(c, oaiResponse)
		} else if strings.HasPrefix(line, "e:") {
			// 完成事件
			var finishData map[string]interface{}
			finishReason := "stop" // 默认值
			if err := json.Unmarshal([]byte(line[2:]), &finishData); err == nil {
				if reason, ok := finishData["finishReason"].(string); ok {
					finishReason = reason
				}
			}
			oaiResponse.Choices = models.NewChoice("", "", finishReason)
			h.writeStreamResponse(c, oaiResponse)
		}
	}
	c.Writer.Write([]byte("data: [DONE]\n\n"))
	c.Writer.Flush()
}

func (h *ChatHandler) writeStreamResponse(c *gin.Context, oaiResponse *models.OpenAIChatCompletionsStreamResponse) {
	headerJSON, _ := json.Marshal(oaiResponse)
	c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", headerJSON)))
	c.Writer.Flush()
}

// 处理常规响应
func (h *ChatHandler) handleRegularResponse(c *gin.Context, resp *req.Response, model string) {
	defer resp.Body.Close()
	c.Header("Content-Type", "application/json")
	c.Header("Access-Control-Allow-Origin", "*")

	scanner := bufio.NewScanner(resp.Body)

	var content, reasoningContent string
	usage := models.Usage{}
	finishReason := "stop"

	clientDone := c.Request.Context().Done()
	for scanner.Scan() {
		select {
		case <-clientDone:
			return
		default:
			//do nothing
		}
		line := scanner.Text()
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "0:") {
			// 内容部分
			content += processContent(line[2:])
		} else if strings.HasPrefix(line, "g:") {
			// 推理内容
			reasoningContent += processContent(line[2:])
		} else if strings.HasPrefix(line, "e:") {
			// 完成信息
			var finishData map[string]interface{}
			if err := json.Unmarshal([]byte(line[2:]), &finishData); err == nil {
				if reason, ok := finishData["finishReason"].(string); ok {
					finishReason = reason
				}
			}
		} else if strings.HasPrefix(line, "d:") {
			// 用量信息
			var usageData map[string]interface{}
			if err := json.Unmarshal([]byte(line[2:]), &usageData); err == nil {
				if u, ok := usageData["usage"].(map[string]interface{}); ok {
					if pt, ok := u["promptTokens"].(float64); ok {
						usage.PromptTokens = int(pt)
					}
					if ct, ok := u["completionTokens"].(float64); ok {
						usage.CompletionTokens = int(ct)
					}
					usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
				}
			}
		}
	}

	// 构造OpenAI格式的响应
	id := fmt.Sprintf("chatcmpl-%s%s", time.Now().Format("20060102150405"), randString(10))

	choices := models.NewChoice(content, reasoningContent, finishReason)
	oaiResponse := models.NewOaiStreamResponse(id, time.Now().Unix(), model, choices)
	oaiResponse.Usage = usage

	c.JSON(http.StatusOK, oaiResponse)
}

// 辅助函数：处理内容，移除引号并处理转义
func processContent(s string) string {
	// 移除开头和结尾的引号
	s = strings.TrimPrefix(s, "\"")
	s = strings.TrimSuffix(s, "\"")

	// 处理转义的换行符
	s = strings.ReplaceAll(s, "\\n", "\n")

	return s
}

// 辅助函数：生成随机字符串
func randString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
