package llm

import (
	"strings"
	"testing"
)

func TestBuildBundleUsesSharedPrompt(t *testing.T) {
	queries := []CriterionQuery{
		{ID: "1.4.4", Name: "Resize Text", Level: "AA", Requirement: "Text can resize to 200%."},
	}
	page := PageContext{URL: "https://example.com", Title: "Example", Language: "en"}

	b := BuildBundle(queries, page)
	if b.SystemPrompt != CriterionSystemPrompt {
		t.Errorf("bundle system prompt differs from the shared prompt")
	}
	if len(b.Items) != 1 || b.Items[0].Query.ID != "1.4.4" {
		t.Fatalf("unexpected bundle items: %+v", b.Items)
	}
	if !strings.Contains(b.Items[0].Prompt, "1.4.4 Resize Text") {
		t.Errorf("prompt missing criterion; got:\n%s", b.Items[0].Prompt)
	}
	// No screenshot → the prompt must warn against guessing visual criteria.
	if !strings.Contains(b.Items[0].Prompt, "No screenshot") {
		t.Errorf("prompt should note missing screenshot")
	}
}

func TestParseVerdicts(t *testing.T) {
	// Tolerates a markdown fence, as an agent or model might emit.
	data := "```json\n[{\"id\":\"1.2.1\",\"conformance\":\"Not Applicable\",\"confidence\":0.9,\"reasoning\":\"no media\"}]\n```"
	verdicts, err := ParseVerdicts([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if len(verdicts) != 1 || verdicts[0].ID != "1.2.1" || verdicts[0].Conformance != VerdictNotApplicable {
		t.Fatalf("unexpected verdicts: %+v", verdicts)
	}
}

func TestParseVerdictFlexibleAcceptsBareObject(t *testing.T) {
	vs, err := parseVerdictFlexible(`{"conformance":"Supports","confidence":0.8,"reasoning":"ok"}`, "2.4.4")
	if err != nil {
		t.Fatal(err)
	}
	if len(vs) != 1 || vs[0].ID != "2.4.4" || vs[0].Conformance != VerdictSupports {
		t.Fatalf("bare object not normalized: %+v", vs)
	}
}
