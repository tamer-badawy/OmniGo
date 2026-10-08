// Copyright 2026 Tamer Badawy
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package ai

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	cohere "github.com/cohere-ai/cohere-go/v2"
	"github.com/cohere-ai/cohere-go/v2/client"
)

type CohereClient struct {
	model       string
	apiKey      string
	description string

	client      *client.Client
	chatSession []Message
}

func NewCohereClient(ctx context.Context, model, description, apiKey string) (*CohereClient, error) {

	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("API Key is required for Cohere, update it on Settings")
	}

	client := client.NewClient(client.WithToken(apiKey))

	return &CohereClient{
		model:       model,
		description: description,
		apiKey:      apiKey,
		client:      client,
		chatSession: []Message{},
	}, nil
}

func (c *CohereClient) GetName() string {
	return c.model
}

func (c *CohereClient) GetDescription() string {
	return c.description
}

func (c *CohereClient) GetAPIKey() string {
	return c.apiKey
}

func (c *CohereClient) SetAPIKey(apiKey string) {
	c.apiKey = apiKey
}

func (c *CohereClient) ClearSession() {
	c.chatSession = []Message{}
}

func (c *CohereClient) GenerateResponse(ctx context.Context, prompt string, attachmentPath []string, onTokenChunk func(string)) error {
	var contents []*cohere.Content
	var historyContents []*cohere.Content
	var assistantContents []*cohere.AssistantMessageV2ContentOneItem

	if strings.TrimSpace(prompt) != "" {
		contents = append(contents, &cohere.Content{
			Text: &cohere.ChatTextContent{
				Text: prompt,
			},
		})
	}
	for _, path := range attachmentPath {
		fileBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read the attachment from the disk: %w", err)
		}
		mimeType := mime.TypeByExtension(filepath.Ext(path))
		if mimeType == "" {
			mimeType = "application/octet-stream" // Default MIME type if unknown
		}
		imageBase64 := base64.StdEncoding.EncodeToString(fileBytes)
		dataUrl := fmt.Sprintf("data:%s;base64,%s", mimeType, imageBase64)
		contents = append(contents, &cohere.Content{
			ImageUrl: &cohere.ImageContent{
				ImageUrl: &cohere.ImageUrl{
					Url: dataUrl,
				},
			},
		})
	}

	if len(contents) == 0 {
		return fmt.Errorf("no prompt or attachments provided for response generation")
	}

	// attach history message as cohere don't keep track of the conversation history, so we need to send it with every request

	for _, msg := range c.chatSession {
		switch msg.Role {
		case "user":
			historyContents = append(historyContents, &cohere.Content{
				Text: &cohere.ChatTextContent{
					Text: msg.Text,
				},
			})
			for _, path := range msg.Attachments {
				fileBytes, err := os.ReadFile(path)
				if err != nil {
					return fmt.Errorf("failed to read the attachment from the disk: %w", err)
				}
				mimeType := mime.TypeByExtension(filepath.Ext(path))
				if mimeType == "" {
					mimeType = "application/octet-stream" // Default MIME type if unknown
				}
				imageBase64 := base64.StdEncoding.EncodeToString(fileBytes)
				dataUrl := fmt.Sprintf("data:%s;base64,%s", mimeType, imageBase64)
				historyContents = append(historyContents, &cohere.Content{
					ImageUrl: &cohere.ImageContent{
						ImageUrl: &cohere.ImageUrl{
							Url: dataUrl,
						},
					},
				})
			}
		case "model":
			assistantContents = append(assistantContents, &cohere.AssistantMessageV2ContentOneItem{
				Text: &cohere.ChatTextContent{
					Text: msg.Text,
				},
			})
		}
	}

	systemMessage := &cohere.SystemMessageV2{
		Content: &cohere.SystemMessageV2Content{
			SystemMessageV2ContentOneItemList: []*cohere.SystemMessageV2ContentOneItem{{
				Text: &cohere.ChatTextContent{
					Text: SystemRule,
				},
			}},
		},
	}

	userMessage := &cohere.UserMessageV2{
		Content: &cohere.UserMessageV2Content{
			ContentList: contents,
		},
	}

	var allMessages []*cohere.ChatMessageV2
	var historyMessages []*cohere.ChatMessageV2

	allMessages = append(allMessages, &cohere.ChatMessageV2{
		System: systemMessage,
	})

	if len(assistantContents) > 0 {
		historyMessages = append(historyMessages, &cohere.ChatMessageV2{
			User: &cohere.UserMessageV2{
				Content: &cohere.UserMessageV2Content{
					ContentList: historyContents,
				},
			},
		})
		historyMessages = append(historyMessages, &cohere.ChatMessageV2{
			Assistant: &cohere.AssistantMessage{
				Content: &cohere.AssistantMessageV2Content{
					AssistantMessageV2ContentOneItemList: assistantContents,
				},
			},
		})
		allMessages = append(allMessages, historyMessages...)
	}
	allMessages = append(allMessages, &cohere.ChatMessageV2{
		User: userMessage,
	})

	stream, err := c.client.V2.ChatStream(ctx, &cohere.V2ChatStreamRequest{
		Model:    c.model,
		Messages: allMessages,
	})
	if err != nil {
		return err
	}
	defer stream.Close()

	// Save the user message to the chat session history

	//check if the length of the chat session is greater than 16, if so remove the first two message to keep the history size manageable
	if len(c.chatSession) == 16 {
		c.chatSession = c.chatSession[2:]
	}
	c.chatSession = append(c.chatSession, Message{
		Role:      "user",
		Text:      prompt,
		Timestamp: time.Now(),
	})
	textBuffer := ""
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			// End of stream
			c.chatSession = append(c.chatSession, Message{
				Role:      "model",
				Text:      textBuffer,
				Timestamp: time.Now(),
			})
			break
		}

		if err != nil {
			return fmt.Errorf("error while receiving token chunk: %w", err)
		}
		if response.ContentDelta != nil && response.ContentDelta.Delta != nil &&
			response.ContentDelta.Delta.Message != nil && response.ContentDelta.Delta.Message.Content != nil && response.ContentDelta.Delta.Message.Content.Text != nil {
			onTokenChunk(*response.ContentDelta.Delta.Message.Content.Text)
			textBuffer += *response.ContentDelta.Delta.Message.Content.Text
		}
	}

	return nil

}

func (c *CohereClient) LoadHistory(messages []Message) error {
	c.chatSession = messages
	return nil
}
