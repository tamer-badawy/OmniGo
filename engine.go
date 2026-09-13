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
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"github.com/tamer-badawy/OmniGo/ai"
	"github.com/tamer-badawy/OmniGo/ui"
)

type AIProviderInfo struct {
	Name        string
	APIKeyName  string
	Description string
	Model       string
}

var AIProviders = []AIProviderInfo{
	{Name: "Gemini", APIKeyName: "gemini_key", Description: "Google Gemini AI Client", Model: "Gemini 3.6 Flash"},
	{Name: "OpenAI", APIKeyName: "openai_key", Description: "OpenAI ChatGPT AI Client", Model: "ChatGPT-4o"},
}

func GetAIProvidersName() []string {
	names := make([]string, len(AIProviders))
	for i, v := range AIProviders {
		names[i] = v.Name
	}
	return names
}

func (app *OmniApp) SetActiveAIClient() error {
	// OmniApp.ActiveAIProvider is Already Set
	apiKey := app.Preferences.StringWithFallback(app.ActiveAIProvider.APIKeyName, "")
	var err error
	switch app.ActiveAIProvider.Name {
	case "Gemini":

		app.ActiveAI, err = ai.NewGeminiClient(app.context, app.ActiveAIProvider.Model, app.ActiveAIProvider.Description, apiKey)
		return err
	case "OpenAI":
		app.ActiveAIProvider = AIProviders[1]

	}
	// No Implemented AIProvider found
	return fmt.Errorf("failed No AI Provider found or not implemented : %s", app.ActiveAIProvider.Name)

}

func (app *OmniApp) LoadSavedAIProvider() {
	aiProvider := app.Preferences.StringWithFallback("AIProvider", AIProviders[0].Name)
	selectedProviderIndex := slices.IndexFunc(AIProviders, func(p AIProviderInfo) bool {
		return p.Name == aiProvider
	})
	if selectedProviderIndex == -1 {
		app.ActiveAIProvider = AIProviders[0]
	} else {
		app.ActiveAIProvider = AIProviders[selectedProviderIndex]
	}
	app.AISelector.SetSelected(app.ActiveAIProvider.Name)
}

func (app *OmniApp) HandlePromptSubmission() {

	userPrompt := app.PromptField.Text
	if strings.TrimSpace(userPrompt) == "" {
		return
	}
	// Comment out the lines between the "Benchmarking Speed Test" comments to disable benchmarking
	// Benchmarking Speed Test
	pipelineStart := time.Now()
	var firstTokenTime time.Time
	firstTokenReceived := false
	characterCount := 0
	fmt.Println("\n⏱️ Benchmarking Speed Test Started...")
	// Benchmarking Speed Test

	var aiBubble *ui.ChatBubble

	fyne.Do(func() {
		app.PromptField.SetText("")
		bubble := ui.NewChatBubble("user", userPrompt, app.ActiveAIProvider.Model)
		app.ChatLogContainer.Add(bubble)
		app.ScrollContainer.ScrollToBottom()
		aiBubble = ui.NewChatBubble("ai", "", app.ActiveAIProvider.Model)
		app.ChatLogContainer.Add(aiBubble)
		aiBubble.Thinking()
		app.ScrollContainer.ScrollToBottom()
	})
	// Check if this is the first message the  the  session , and  change the  title
	if len(app.CurrentSession.Messages) == 0 {
		titleLen := len(userPrompt)
		if titleLen > 60 {
			titleLen = 60
		}
		app.CurrentSession.Title = userPrompt[:titleLen] + "..."
		fyne.Do(func() {
			app.SidebarList.Refresh()
		})
	}
	app.CurrentSession.Messages = append(app.CurrentSession.Messages, ai.Message{
		Role:      "user",
		Text:      userPrompt,
		Timestamp: time.Now(),
	})
	errChan := make(chan error, 1)
	go func() {
		if app.ActiveAI == nil {
			errChan <- fmt.Errorf("AI client is not initialized. Please check your API keys and settings.")
			return
		}
		aiTextBuffer := ""
		err := app.ActiveAI.GenerateResponse(context.Background(), userPrompt, nil, func(tokenChunk string) {
			// Benchmarking Speed Test
			if !firstTokenReceived {
				firstTokenTime = time.Now()
				firstTokenReceived = true
				ttft := firstTokenTime.Sub(pipelineStart)
				fmt.Printf("⏱️ Time to First Token (TTFT): %v\n", ttft)
			}
			// End of time to first token benchmark
			// Count the number of characters received in the token chunk
			characterCount += len(tokenChunk)
			// Benchmarking Speed Test

			fyne.Do(func() {
				aiTextBuffer += tokenChunk
				aiBubble.UpdateText(aiTextBuffer)
				app.ScrollContainer.ScrollToBottom()
			})
		})
		errChan <- err

		// Benchmarking Speed Test
		pipelineEnd := time.Now()
		pipelineDuration := pipelineEnd.Sub(pipelineStart)

		streamDuration := pipelineEnd.Sub(firstTokenTime)
		charsPerSecond := float64(characterCount) / streamDuration.Seconds()
		fmt.Println("==================================================")
		fmt.Printf("🏁 [BENCHMARK REPORT] Session Completed Successfully\n")

		fmt.Printf("⏱️  Total Turnaround Time:   %v\n", pipelineDuration)
		fmt.Printf("⏳ Network Stream Duration:  %v\n", streamDuration)
		fmt.Printf("🔤 Total Characters Recv:  %d chars\n", characterCount)
		fmt.Printf("🚀 App Throughput Speed:    %.2f chars/sec\n", charsPerSecond)
		fmt.Println("==================================================")

		// Benchmarking Speed Test

		app.CurrentSession.Messages = append(app.CurrentSession.Messages, ai.Message{
			Role:      "model",
			Text:      aiTextBuffer,
			Timestamp: time.Now(),
		})

		app.StorageManager.SaveSession(app.CurrentSession)

	}()

	if err := <-errChan; err != nil {
		aiBubble.UpdateText(fmt.Sprintf("❌ **Error:** %v", err))
	}

}

func (app *OmniApp) LoadAndSynchronizeHistoryOnLaunch() {
	records, err := app.StorageManager.LoadAllSessions()
	if err != nil || len(records) == 0 {
		initialSession := &ChatSession{
			ID:        fmt.Sprintf("session_%d", time.Now().UnixNano()),
			Messages:  []ai.Message{},
			Title:     "New Conversation Thread",
			UpdatedAt: time.Now(),
		}
		app.SessionList = append(app.SessionList, initialSession)
		app.CurrentSession = initialSession
		app.SidebarList.Refresh()
		app.SidebarList.Select(0)
		return
	}
	app.SessionList = records
	app.SidebarList.Refresh()
	app.SidebarList.Select(0)
}

func (app *OmniApp) SwitchChatSession(index int) {
	if index < 0 || index >= len(app.SessionList) {
		return
	}
	app.CurrentSession = app.SessionList[index]
	app.ActiveAI.LoadHistory(app.CurrentSession.Messages)

	// Update the  UI
	go func() {
		fyne.Do(func() {
			app.ChatLogContainer.Objects = nil
			app.ChatLogContainer.Refresh()
			for _, msg := range app.CurrentSession.Messages {
				bubble := ui.NewChatBubble(msg.Role, msg.Text, app.ActiveAIProvider.Model)
				app.ChatLogContainer.Add(bubble)
			}
			app.ChatLogContainer.Refresh()
			app.ScrollContainer.ScrollToBottom()
		})
	}()
}
