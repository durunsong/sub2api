package securityaudit

import (
	"context"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEphemeralAuditStillEvaluatesWithoutDurableQueue(t *testing.T) {
	prompt := &fakePromptEngine{mode: ModeAsync, decision: &PromptDecision{Kind: DecisionBlock}}
	legacy := &fakeLegacyEngine{decision: &LegacyDecision{Allowed: true}}
	decision := NewCoordinator(legacy, prompt).Check(ctxkey.WithNoContentPersistence(context.Background()), Request{Body: []byte(`{"prompt":"private"}`)})
	require.True(t, decision.AllowNextStage, "observe mode must remain non-blocking")
	require.Equal(t, int64(1), prompt.evaluates.Load())
	require.Zero(t, prompt.enqueues.Load())
	require.Equal(t, int64(1), legacy.calls.Load())
}
func TestEphemeralBlockingAuditDoesNotPersistContent(t *testing.T) {
	repo := &fakeJobRepository{}
	evaluator := newGuardEvaluator(&scriptedScanner{}, repo, nil, 4, 2)
	result, err := evaluator.Evaluate(ctxkey.WithNoContentPersistence(context.Background()), guardConfig(ActiveEndpoint{ID: "good", Enabled: true, TimeoutMS: 1000, InputLimit: 100}), PromptSnapshot{RequestID: "r", ScanText: "private", FullPrompt: "private", RedactedPreview: "private"})
	require.NoError(t, err)
	require.Equal(t, DecisionAllow, result.Kind)
	require.Zero(t, repo.recordBlockingCalls)
}

func TestEphemeralPromptServiceScansConfiguredObserveMode(t *testing.T) {
	repo := &fakeJobRepository{}
	var scanned string
	scanner := PromptScannerFunc(func(_ context.Context, _ ActiveEndpoint, chunk string, _ []string) (*NormalizedResult, error) {
		scanned = chunk
		return &NormalizedResult{Decision: EventPass, RiskLevel: RiskLow, Action: ActionAllow, ScannerScores: map[string]float64{}, ScannerEvidence: map[string]string{}}, nil
	})
	svc := &PromptService{
		config: &fakeConfigStore{active: true, cfg: ActiveConfig{
			RiskControlEnabled: true, Enabled: true, AllGroups: true,
			Scanners: AllScannerIDs, Endpoints: []ActiveEndpoint{{ID: "guard", Enabled: true, TimeoutMS: 1000, InputLimit: 4096}},
		}},
		evaluator: newGuardEvaluator(scanner, repo, nil, 2, 2),
	}
	decision := NewCoordinator(nil, svc).Check(ctxkey.WithNoContentPersistence(context.Background()), Request{Protocol: "openai_images", Body: []byte(`{"prompt":"private creative prompt"}`)})
	require.True(t, decision.AllowNextStage)
	require.Equal(t, "private creative prompt", scanned)
	require.Zero(t, repo.recordBlockingCalls)
}
