package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResponsesUsageCacheWriteTTL(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name               string
		body               string
		total, short, long int
	}{
		{"mixed", `{"input_tokens_details":{"cache_write_tokens":20,"cache_write_tokens_details":{"ttl_30m_tokens":2,"ttl_12h_tokens":18}}}`, 20, 2, 18},
		{"chat aliases", `{"prompt_tokens_details":{"cache_write_tokens":20,"cache_write_tokens_details":{"ttl_30m_tokens":2,"ttl_12h_tokens":18}}}`, 20, 2, 18},
		{"canonical zero", `{"cache_creation_input_tokens":20,"input_tokens_details":{"cache_write_tokens":0,"cache_write_tokens_details":{"ttl_30m_tokens":2,"ttl_12h_tokens":18}}}`, 0, 0, 0},
		{"missing details", `{"input_tokens_details":{"cache_write_tokens":20}}`, 20, 0, 0},
		{"partial details", `{"input_tokens_details":{"cache_write_tokens":20,"cache_write_tokens_details":{"ttl_12h_tokens":20}}}`, 20, 0, 0},
		{"strings", `{"input_tokens_details":{"cache_write_tokens":20,"cache_write_tokens_details":{"ttl_30m_tokens":"2","ttl_12h_tokens":18}}}`, 20, 0, 0},
		{"fractions", `{"input_tokens_details":{"cache_write_tokens":20,"cache_write_tokens_details":{"ttl_30m_tokens":2.5,"ttl_12h_tokens":17.5}}}`, 20, 0, 0},
		{"negative", `{"input_tokens_details":{"cache_write_tokens":20,"cache_write_tokens_details":{"ttl_30m_tokens":-2,"ttl_12h_tokens":22}}}`, 20, 0, 0},
		{"inconsistent sum", `{"input_tokens_details":{"cache_write_tokens":20,"cache_write_tokens_details":{"ttl_30m_tokens":2,"ttl_12h_tokens":17}}}`, 20, 0, 0},
		{"untrusted top-level buckets", `{"cache_creation_input_tokens":20,"cache_creation_30m_input_tokens":2,"cache_creation_12h_input_tokens":18}`, 20, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Reusing a value must not retain TTL buckets from an earlier response.
			usage := ResponsesUsage{CacheCreation30mInputTokens: 5, CacheCreation12hInputTokens: 15}
			require.NoError(t, json.Unmarshal([]byte(tc.body), &usage))
			require.Equal(t, tc.total, usage.CacheCreationInputTokens)
			require.Equal(t, tc.short, usage.CacheCreation30mInputTokens)
			require.Equal(t, tc.long, usage.CacheCreation12hInputTokens)
		})
	}
}
