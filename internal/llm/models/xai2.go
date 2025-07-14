package models

const (
	// Provider name - using xai2 to distinguish from the existing xai
	ProviderXAI2 ModelProvider = "xai2"
)

// xAI Model IDs - All prefixed with xai2. for complete independence
const (
	// Grok 4 - Latest flagship model with 256k context window
	XAI2Grok4     ModelID = "xai2.grok-4"
	XAI2Grok40709 ModelID = "xai2.grok-4-0709" // Full model name
	
	// Grok 3 - Enterprise model with 131k context window
	XAI2Grok3     ModelID = "xai2.grok-3"
	XAI2Grok3Beta ModelID = "xai2.grok-3-beta"
	
	// Grok 3 Mini - Lightweight reasoning model
	XAI2Grok3Mini     ModelID = "xai2.grok-3-mini"
	XAI2Grok3MiniBeta ModelID = "xai2.grok-3-mini-beta"
	
	// Fast variants
	XAI2Grok3Fast     ModelID = "xai2.grok-3-fast"
	XAI2Grok3FastBeta ModelID = "xai2.grok-3-fast-beta"
	XAI2Grok3MiniFast ModelID = "xai2.grok-3-mini-fast"
	XAI2Grok3MiniFastBeta ModelID = "xai2.grok-3-mini-fast-beta"
	
	// Grok 2 Vision - Multimodal model
	XAI2Grok2Vision     ModelID = "xai2.grok-2-vision"
	XAI2Grok2Vision1212 ModelID = "xai2.grok-2-vision-1212" // Full model name
	
	// Grok 2 Image - Image generation model
	XAI2Grok2Image     ModelID = "xai2.grok-2-image"
	XAI2Grok2Image1212 ModelID = "xai2.grok-2-image-1212" // Full model name
)

var XAI2Models = map[ModelID]Model{
	// Grok 4 - Flagship model
	XAI2Grok4: {
		ID:                 XAI2Grok4,
		Name:               "Grok 4",
		Provider:           ProviderXAI2,
		APIModel:           "grok-4-0709", // Use full name as primary
		CostPer1MIn:        3.0,
		CostPer1MInCached:  0.75,
		CostPer1MOut:       15.0,
		CostPer1MOutCached: 0,
		ContextWindow:      256_000, // 256k context window
		DefaultMaxTokens:   20_000,
		CanReason:           true,
		SupportsAttachments: true, // Supports function calling
	},
	XAI2Grok40709: {
		ID:                 XAI2Grok40709,
		Name:               "Grok 4 (0709)",
		Provider:           ProviderXAI2,
		APIModel:           "grok-4-0709",
		CostPer1MIn:        3.0,
		CostPer1MInCached:  0.75,
		CostPer1MOut:       15.0,
		CostPer1MOutCached: 0,
		ContextWindow:      256_000,
		DefaultMaxTokens:   20_000,
		CanReason:           true,
		SupportsAttachments: true, // Supports function calling
	},
	
	// Grok 3 - Enterprise model
	XAI2Grok3: {
		ID:                 XAI2Grok3,
		Name:               "Grok 3",
		Provider:           ProviderXAI2,
		APIModel:           "grok-3",
		CostPer1MIn:        3.0,
		CostPer1MInCached:  0.75,
		CostPer1MOut:       15.0,
		CostPer1MOutCached: 0,
		ContextWindow:      131_072,
		DefaultMaxTokens:   20_000,
		CanReason:           true,
		SupportsAttachments: true, // Supports function calling
	},
	XAI2Grok3Beta: {
		ID:                 XAI2Grok3Beta,
		Name:               "Grok 3 Beta",
		Provider:           ProviderXAI2,
		APIModel:           "grok-3-beta",
		CostPer1MIn:        3.0,
		CostPer1MInCached:  0.75,
		CostPer1MOut:       15.0,
		CostPer1MOutCached: 0,
		ContextWindow:      131_072,
		DefaultMaxTokens:   20_000,
		CanReason:           true,
		SupportsAttachments: true, // Supports function calling
	},
	
	// Grok 3 Mini - Lightweight model
	XAI2Grok3Mini: {
		ID:                 XAI2Grok3Mini,
		Name:               "Grok 3 Mini",
		Provider:           ProviderXAI2,
		APIModel:           "grok-3-mini",
		CostPer1MIn:        0.30,
		CostPer1MInCached:  0.075,
		CostPer1MOut:       0.50,
		CostPer1MOutCached: 0,
		ContextWindow:      131_072,
		DefaultMaxTokens:   20_000,
		CanReason:           true,
		SupportsAttachments: true, // Supports function calling
	},
	XAI2Grok3MiniBeta: {
		ID:                 XAI2Grok3MiniBeta,
		Name:               "Grok 3 Mini Beta",
		Provider:           ProviderXAI2,
		APIModel:           "grok-3-mini-beta",
		CostPer1MIn:        0.30,
		CostPer1MInCached:  0.075,
		CostPer1MOut:       0.50,
		CostPer1MOutCached: 0,
		ContextWindow:      131_072,
		DefaultMaxTokens:   20_000,
		CanReason:           true,
		SupportsAttachments: true, // Supports function calling
	},
	
	// Fast variants - From web search, with estimated pricing
	XAI2Grok3Fast: {
		ID:                 XAI2Grok3Fast,
		Name:               "Grok 3 Fast",
		Provider:           ProviderXAI2,
		APIModel:           "grok-3-fast",
		CostPer1MIn:        5.0,  // Higher cost for faster response
		CostPer1MInCached:  1.25,
		CostPer1MOut:       25.0,
		CostPer1MOutCached: 0,
		ContextWindow:      131_072,
		DefaultMaxTokens:   20_000,
		CanReason:           true,
		SupportsAttachments: true, // Supports function calling
	},
	XAI2Grok3FastBeta: {
		ID:                 XAI2Grok3FastBeta,
		Name:               "Grok 3 Fast Beta",
		Provider:           ProviderXAI2,
		APIModel:           "grok-3-fast-beta",
		CostPer1MIn:        5.0,
		CostPer1MInCached:  1.25,
		CostPer1MOut:       25.0,
		CostPer1MOutCached: 0,
		ContextWindow:      131_072,
		DefaultMaxTokens:   20_000,
		CanReason:           true,
		SupportsAttachments: true, // Supports function calling
	},
	XAI2Grok3MiniFast: {
		ID:                 XAI2Grok3MiniFast,
		Name:               "Grok 3 Mini Fast",
		Provider:           ProviderXAI2,
		APIModel:           "grok-3-mini-fast",
		CostPer1MIn:        0.60,
		CostPer1MInCached:  0.15,
		CostPer1MOut:       4.0,
		CostPer1MOutCached: 0,
		ContextWindow:      131_072,
		DefaultMaxTokens:   20_000,
		CanReason:           true,
		SupportsAttachments: true, // Supports function calling
	},
	XAI2Grok3MiniFastBeta: {
		ID:                 XAI2Grok3MiniFastBeta,
		Name:               "Grok 3 Mini Fast Beta",
		Provider:           ProviderXAI2,
		APIModel:           "grok-3-mini-fast-beta",
		CostPer1MIn:        0.60,
		CostPer1MInCached:  0.15,
		CostPer1MOut:       4.0,
		CostPer1MOutCached: 0,
		ContextWindow:      131_072,
		DefaultMaxTokens:   20_000,
		CanReason:           true,
		SupportsAttachments: true, // Supports function calling
	},
	
	// Grok 2 Vision - Multimodal model
	XAI2Grok2Vision: {
		ID:                 XAI2Grok2Vision,
		Name:               "Grok 2 Vision",
		Provider:           ProviderXAI2,
		APIModel:           "grok-2-vision-1212", // Use full name as primary
		CostPer1MIn:        2.0,
		CostPer1MInCached:  0, // No cached input pricing for vision
		CostPer1MOut:       10.0,
		CostPer1MOutCached: 0,
		ContextWindow:      32_768, // 32k context window
		DefaultMaxTokens:   8_000,
		CanReason:           true,
		SupportsAttachments: true, // Supports function calling and vision
	},
	XAI2Grok2Vision1212: {
		ID:                 XAI2Grok2Vision1212,
		Name:               "Grok 2 Vision (1212)",
		Provider:           ProviderXAI2,
		APIModel:           "grok-2-vision-1212",
		CostPer1MIn:        2.0,
		CostPer1MInCached:  0,
		CostPer1MOut:       10.0,
		CostPer1MOutCached: 0,
		ContextWindow:      32_768,
		DefaultMaxTokens:   8_000,
		CanReason:           true,
		SupportsAttachments: true, // Supports function calling and vision
	},
	
	// Grok 2 Image - Image generation model
	XAI2Grok2Image: {
		ID:                 XAI2Grok2Image,
		Name:               "Grok 2 Image",
		Provider:           ProviderXAI2,
		APIModel:           "grok-2-image-1212", // Use full name as primary
		CostPer1MIn:        0, // Image generation doesn't use token pricing
		CostPer1MInCached:  0,
		CostPer1MOut:       0,
		CostPer1MOutCached: 0,
		// Note: This is an image generation model with $0.07 per image pricing
		ContextWindow:      0, // Not applicable for image generation
		DefaultMaxTokens:   0, // Not applicable for image generation
		// Image generation model - no text capabilities
	},
	XAI2Grok2Image1212: {
		ID:                 XAI2Grok2Image1212,
		Name:               "Grok 2 Image (1212)",
		Provider:           ProviderXAI2,
		APIModel:           "grok-2-image-1212",
		CostPer1MIn:        0,
		CostPer1MInCached:  0,
		CostPer1MOut:       0,
		CostPer1MOutCached: 0,
		// Note: $0.07 per image pricing
		ContextWindow:      0,
		DefaultMaxTokens:   0,
		// Image generation model - no text capabilities
	},
}

// Model aliases mapping - maps aliases to primary model IDs
// All aliases are also prefixed with xai2. for complete independence
var XAI2ModelAliases = map[string]ModelID{
	// Grok 4 aliases
	"xai2.grok-4":        XAI2Grok40709, // Maps to full model name
	"xai2.grok-4-latest": XAI2Grok40709,
	
	// Grok 3 aliases
	"xai2.grok-3-latest": XAI2Grok3,
	"xai2.grok-3-beta":   XAI2Grok3Beta,
	
	// Grok 3 Fast aliases
	"xai2.grok-3-fast-latest": XAI2Grok3Fast,
	"xai2.grok-3-fast-beta":   XAI2Grok3FastBeta,
	
	// Grok 3 Mini aliases
	"xai2.grok-3-mini-latest": XAI2Grok3Mini,
	"xai2.grok-3-mini-beta":   XAI2Grok3MiniBeta,
	
	// Grok 3 Mini Fast aliases
	"xai2.grok-3-mini-fast-latest": XAI2Grok3MiniFast,
	"xai2.grok-3-mini-fast-beta":   XAI2Grok3MiniFastBeta,
	
	// Grok 2 Vision aliases
	"xai2.grok-2-vision":        XAI2Grok2Vision1212, // Maps to full model name
	"xai2.grok-2-vision-latest": XAI2Grok2Vision1212,
	
	// Grok 2 Image aliases
	"xai2.grok-2-image":        XAI2Grok2Image1212, // Maps to full model name
	"xai2.grok-2-image-latest": XAI2Grok2Image1212,
}