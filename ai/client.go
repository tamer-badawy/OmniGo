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
	"time"
)

type Message struct {
	Role      string    `json:"role"` // user or model
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

type AIClient interface {
	// GetName returns the name of the AI client.
	GetName() string

	// GetDescription returns a brief description of the AI client.
	GetDescription() string

	// GetIcon returns the icon associated with the AI client.
	// GetIcon() []byte

	// GetAPIKey returns the API key used for authentication with the AI service.
	GetAPIKey() string

	// SetAPIKey sets the API key for authentication with the AI service.
	SetAPIKey(key string)

	//GetHistory return the  converstion history of the AI client.
	//GetHistory() []Message

	//LoadHistory poplute the provider's chat with old message.
	LoadHistory(messages []Message) error

	// GenerateResponse generates a response based on the provided input text and\or attachmentPath.
	GenerateResponse(ctx context.Context, prompt string, attachmentPath []string, onTokenChunk func(string)) error
}
