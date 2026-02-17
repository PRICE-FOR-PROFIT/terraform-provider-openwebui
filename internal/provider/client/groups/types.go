// Copyright (c) HashiCorp, Inc.

package groups

type Group struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Permissions *GroupPermissions      `json:"permissions"`
	Data        map[string]interface{} `json:"data"`
	Meta        map[string]interface{} `json:"meta"`
	UserIDs     []string               `json:"user_ids"`
	CreatedAt   int64                  `json:"created_at"`
	UpdatedAt   int64                  `json:"updated_at"`
}

type GroupPermissions struct {
	Workspace WorkspacePermissions `json:"workspace"`
	Chat      ChatPermissions      `json:"chat"`
	Sharing   SharingPermissions   `json:"sharing"`
	Features  FeaturesPermissions  `json:"features"`
}

type WorkspacePermissions struct {
	Models    bool `json:"models"`
	Knowledge bool `json:"knowledge"`
	Prompts   bool `json:"prompts"`
	Tools     bool `json:"tools"`
}

type ChatPermissions struct {
	FileUpload         bool `json:"file_upload"`
	Delete             bool `json:"delete"`
	Edit               bool `json:"edit"`
	Temporary          bool `json:"temporary"`
	Controls           bool `json:"controls"`
	Valves             bool `json:"valves"`
	SystemPrompt       bool `json:"system_prompt"`
	Params             bool `json:"params"`
	DeleteMessage      bool `json:"delete_message"`
	ContinueResponse   bool `json:"continue_response"`
	RegenerateResponse bool `json:"regenerate_response"`
	RateResponse       bool `json:"rate_response"`
	Share              bool `json:"share"`
	Export             bool `json:"export"`
	Stt                bool `json:"stt"`
	Tts                bool `json:"tts"`
	Call               bool `json:"call"`
	MultipleModels     bool `json:"multiple_models"`
	TemporaryEnforced  bool `json:"temporary_enforced"`
}

type SharingPermissions struct {
	PublicModels    bool `json:"public_models"`
	PublicKnowledge bool `json:"public_knowledge"`
	PublicPrompts   bool `json:"public_prompts"`
	PublicTools     bool `json:"public_tools"`
	PublicNotes     bool `json:"public_notes"`
}

type FeaturesPermissions struct {
	DirectToolServers bool `json:"direct_tool_servers"`
	WebSearch         bool `json:"web_search"`
	ImageGeneration   bool `json:"image_generation"`
	CodeInterpreter   bool `json:"code_interpreter"`
	Notes             bool `json:"notes"`
}
