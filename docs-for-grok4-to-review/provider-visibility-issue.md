# Provider Visibility Issue Analysis

## Problem Statement
Despite having multiple providers configured in ~/.opencode.json with valid API keys, only OpenAI models show up in the model selection dialog.

## Root Cause Analysis

### Configuration Data Flow
```
1. Config File (~/.opencode.json)
   └─> viper.ReadInConfig()
       └─> viper.Unmarshal(&cfg)
           └─> cfg.Providers populated with file data
               └─> Validate() runs
                   └─> getEnabledProviders() filters providers
                       └─> TUI shows filtered list
```

### The Critical Code Path

1. **Config Loading** (`config.Load()`)
   ```go
   cfg = &Config{
       Providers: make(map[models.ModelProvider]Provider),
   }
   // ... viper reads config file ...
   viper.Unmarshal(cfg) // This populates providers from file
   ```

2. **Provider Filtering** (`getEnabledProviders()`)
   ```go
   for providerId, provider := range cfg.Providers {
       if !provider.Disabled {
           providers = append(providers, providerId)
       }
   }
   ```

### Why Providers Might Be Missing

1. **Viper Unmarshal Issue**: The providers in the config file should be unmarshaled, but the key names must match exactly.

2. **Provider Key Format**: The config uses lowercase provider names:
   - Config file: `"anthropic"`, `"gemini"`, `"openai"`, `"xai"`
   - Model constants: `ProviderAnthropic`, `ProviderGemini`, etc.

3. **Possible Case Sensitivity**: Go's json unmarshaling into a map with custom key types might have issues.

## Hypothesis

The issue likely stems from one of these:

1. **Model Provider Type Mismatch**: The `models.ModelProvider` type is a string alias, and the JSON keys might not be matching during unmarshal.

2. **Validation Side Effects**: During validation, providers might be getting removed or modified.

3. **Config File API Key Issue**: The API keys in the config might be invalid, causing providers to be disabled during validation.

## Debug Strategy

1. Check if providers are loaded after unmarshal
2. Check if validation is disabling providers
3. Check the exact provider keys being used
4. Verify API key validation logic

## Solution Approaches

1. **Fix Provider Keys**: Ensure config file keys match the ModelProvider constants
2. **Debug Logging**: Add logging to see which providers are loaded
3. **Skip Validation**: Temporarily bypass provider validation to isolate the issue
4. **Direct Provider Addition**: Manually add providers to the runtime config

## Code Investigation Points

1. `viper.Unmarshal(cfg)` - How does it handle the map keys?
2. `Validate()` - Is it removing/disabling providers?
3. `getEnabledProviders()` - What's the actual content of cfg.Providers?
4. Provider constant definitions - Are they matching the JSON keys?