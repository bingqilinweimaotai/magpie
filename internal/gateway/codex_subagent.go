package gateway

import (
	"context"
	"net/http"

	"github.com/yetone/magpie/internal/provider"
	"github.com/yetone/magpie/internal/usage"
)

// SubagentPick is a Codex subagent's request put on the model the user
// set for Codex's subagents (provider.CodexSubagentModel), in place of
// the one its lead asked for in spawn_agent (willz on Discord). Kept says
// why it went on the model asked for instead, "" when it was put on To.
type SubagentPick struct {
	Asked string `json:"asked"`
	To    string `json:"to"`
	Kept  string `json:"kept,omitempty"`
}

// Moved says the request went on To.
func (p *SubagentPick) Moved() bool { return p != nil && p.Kept == "" }

type subagentPickKey struct{}

func withSubagentPick(r *http.Request, p *SubagentPick) *http.Request {
	if p == nil {
		return r
	}
	return r.WithContext(context.WithValue(r.Context(), subagentPickKey{}, p))
}

func subagentPickOf(r *http.Request) *SubagentPick {
	p, _ := r.Context().Value(subagentPickKey{}).(*SubagentPick)
	return p
}

// codexSubagentPick is what is done with the model of a request: nil when
// it isn't a Codex subagent's or no model is set for them, so the request
// goes on as asked. The model set must still be served by a ChatGPT
// account (a subagent's task is sealed for them) and be one the gateway
// key may use; when it isn't, the request goes on the model asked for and
// the pick says why.
func codexSubagentPick(r *http.Request, agent, kind, asked string) *SubagentPick {
	// Codex by its token or by its User-Agent (fromCodex: its CLI, exec,
	// the IDE extension, the desktop app)
	if agent != "codex" && !fromCodex(r.Header) || usage.PurposeOf(kind) != "kind:collab_spawn" {
		return nil
	}
	to := provider.CodexSubagentModel()
	if to == "" || to == asked {
		return nil
	}
	pick := &SubagentPick{Asked: asked, To: to}
	if why := provider.CodexSubagentUnserved(to); why != "" {
		pick.Kept = why
		return pick
	}
	if who, held := keyHolds(r); held {
		if p, model, ok := provider.Resolve(to); !ok || !modelAllowed(who, p, model) {
			pick.Kept = "the gateway key may not use it"
		}
	}
	return pick
}
