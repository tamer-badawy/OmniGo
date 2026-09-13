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
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/genai"
)

type GeminiClient struct {
	name        string
	description string
	apiKey      string

	client *genai.Client
	chat   *genai.Chat
}

// 🛠️ THE SYSTEM INSTRUCTION:
// This hidden rule forces Gemini to structure its answers perfectly for OmniGo's layout engine
const systemRule = "You are OmniGo, a fast Linux desktop assistant. Follow these strict formatting rules:\n\n" +
	"1. Respond with normal conversational prose for your explanations and commentary.\n\n" +
	"2. All programming scripts, terminal commands, or bash lines MUST be placed inside standard code blocks specifying the language (e.g., ```python or ```bash).\n\n" +
	"3. Any generated emails, text prompts, reusable templates, or copyable paragraph responses MUST be wrapped inside a text code block labeled exactly as ```text."

func NewGeminiClient(ctx context.Context, name, description, apiKey string) (*GeminiClient, error) {

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	chat, err := client.Chats.Create(ctx, "gemini-3.6-flash",
		&genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{{Text: systemRule}},
			},
		},
		nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini chat: %w", err)
	}

	return &GeminiClient{
		name:        name,
		description: description,
		apiKey:      apiKey,
		client:      client,
		chat:        chat,
	}, nil
}

func (g *GeminiClient) GetName() string {
	return g.name
}

func (g *GeminiClient) GetDescription() string {
	return g.description
}

func (g *GeminiClient) GetAPIKey() string {
	return g.apiKey
}

func (g *GeminiClient) SetAPIKey(key string) {
	g.apiKey = key
}

func (g *GeminiClient) GenerateResponse(ctx context.Context, prompt string, attachmentPath []string, onTokenChunk func(string)) error {
	// Implementation for generating response with Gemini AI
	var parts []*genai.Part

	if strings.TrimSpace(prompt) != "" {
		parts = append(parts, &genai.Part{Text: prompt})
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
		filePart := genai.Blob{
			MIMEType: mimeType,
			Data:     fileBytes,
		}
		parts = append(parts, &genai.Part{InlineData: &filePart})
	}

	if len(parts) == 0 {
		return fmt.Errorf("no prompt or attachments provided for response generation")
	}

	stream := g.chat.SendStream(ctx, parts...)

	for chunk, err := range stream {
		if err != nil {
			return fmt.Errorf("error while receiving token chunk: %w", err)
		}
		onTokenChunk(chunk.Text())
	}

	return nil
}

func (g *GeminiClient) LoadHistory(messages []Message) error {

	var sdkHistory []*genai.Content

	for _, m := range messages {
		role := m.Role
		sdkHistory = append(sdkHistory, &genai.Content{
			Role:  role,
			Parts: []*genai.Part{{Text: m.Text}},
		})
	}
	chat, err := g.client.Chats.Create(context.Background(), "gemini-3.6-flash",
		&genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{{Text: systemRule}},
			},
		},
		sdkHistory,
	)

	if err != nil {
		return err
	}
	g.chat = chat
	return nil
}
