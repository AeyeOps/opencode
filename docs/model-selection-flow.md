# Model Selection Flow Deep Dive

## Provider Constants
All provider constants are lowercase strings that match config file keys:
- `ProviderAnthropic = "anthropic"`
- `ProviderOpenAI = "openai"`
- `ProviderGemini = "gemini"`
- `ProviderGROQ = "groq"`
- `ProviderXAI = "xai"`
- etc.

## Model Selection Dialog Flow

### 1. Dialog Initialization (`models.go`)
```go
func (m *modelDialogCmp) setupModels() {
    cfg := config.Get()
    modelInfo := GetSelectedModel(cfg)
    m.availableProviders = getEnabledProviders(cfg)
    m.hScrollPossible = len(m.availableProviders) > 1
    // ...
}
```

### 2. Provider Filtering (`getEnabledProviders`)
```go
func getEnabledProviders(cfg *config.Config) []models.ModelProvider {
    var providers []models.ModelProvider
    for providerId, provider := range cfg.Providers {
        if !provider.Disabled {
            providers = append(providers, providerId)
        }
    }
    // Sort by popularity
    slices.SortFunc(providers, func(a, b models.ModelProvider) int {
        rA := models.ProviderPopularity[a]
        rB := models.ProviderPopularity[b]
        // ...
    })
    return providers
}
```

### 3. Model Filtering (`getModelsForProvider`)
```go
func getModelsForProvider(provider models.ModelProvider) []models.Model {
    var providerModels []models.Model
    for _, model := range models.SupportedModels {
        if model.Provider == provider {
            providerModels = append(providerModels, model)
        }
    }
    // Sort in reverse alphabetical order
    // ...
    return providerModels
}
```

## Potential Issues

### Issue 1: Provider Validation Disabling
During `Validate()`, providers can be disabled if:
```go
for provider, providerCfg := range cfg.Providers {
    if providerCfg.APIKey == "" && !providerCfg.Disabled {
        logging.Warn("provider has no API key, marking as disabled", "provider", provider)
        providerCfg.Disabled = true
        cfg.Providers[provider] = providerCfg
    }
}
```

### Issue 2: API Key Validation
The API keys in the config file might be:
1. Invalid/expired
2. Empty strings after unmarshaling
3. Not properly unmarshaled due to JSON structure

### Issue 3: Config Unmarshaling
Viper might not be properly unmarshaling the nested structure:
```json
{
  "providers": {
    "anthropic": {
      "apiKey": "...",
      "disabled": false
    }
  }
}
```

## Debugging Approach

1. **Check Post-Unmarshal State**: Add logging after `viper.Unmarshal(cfg)` to see what's in `cfg.Providers`

2. **Check Post-Validation State**: Add logging after `Validate()` to see if providers are being disabled

3. **Check API Key Presence**: Log the actual API key values (redacted) to ensure they're not empty

4. **Check Dialog State**: Log the result of `getEnabledProviders()` in the model dialog

## Most Likely Cause

Based on the code analysis, the most likely cause is that during validation, providers are being marked as disabled because:
1. The API keys might be seen as empty after unmarshal
2. The validation logic is too aggressive in disabling providers
3. There's a race condition or state mutation issue

## Next Steps

1. Add debug logging to trace provider state through the config flow
2. Check if the API keys are actually present after unmarshal
3. Verify the validation logic isn't overly restrictive
4. Test with a minimal config to isolate the issue