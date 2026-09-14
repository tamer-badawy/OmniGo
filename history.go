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

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/tamer-badawy/OmniGo/ai"
)

type ChatSession struct {
	ID        string       `json:"id"` // Unique timestamp or UUID identifing string filename
	Messages  []ai.Message `json:"messages"`
	Title     string       `json:"title"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type StorageManager struct {
	DataDir  string
	Provider string // Gemini or OpenAi for now
}

func NewStorageManager(provider string) (*StorageManager, error) {
	baseConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to detect native system config directory: %w", err)
	}
	appHistoryDir := filepath.Join(baseConfigDir, "OmniGo", "history", provider)

	if err := os.MkdirAll(appHistoryDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to instantiate history database layout: %w", err)
	}

	return &StorageManager{
		DataDir:  appHistoryDir,
		Provider: provider,
	}, nil
}

func (sm *StorageManager) SaveSession(session *ChatSession) error {
	session.UpdatedAt = time.Now()

	jsonByte, err := json.MarshalIndent(session, "", " ")
	if err != nil {
		return err
	}
	filePath := filepath.Join(sm.DataDir, fmt.Sprintf("%s.json", session.ID))
	return os.WriteFile(filePath, jsonByte, 0644)
}

func (sm *StorageManager) DeleteSession(session *ChatSession) error {
	filePath := filepath.Join(sm.DataDir, fmt.Sprintf("%s.json", session.ID))
	return os.Remove(filePath)
}

func (sm *StorageManager) LoadAllSessions() ([]*ChatSession, error) {
	files, err := os.ReadDir(sm.DataDir)
	if err != nil {
		return nil, err
	}
	var sessions []*ChatSession
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}
		filePath := filepath.Join(sm.DataDir, file.Name())
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		var session ChatSession
		if err := json.Unmarshal(fileBytes, &session); err == nil {
			sessions = append(sessions, &session)
		}

	}
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})
	return sessions, nil
}
