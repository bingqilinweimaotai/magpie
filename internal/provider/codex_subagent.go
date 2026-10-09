package provider

import (
	"fmt"
	"strings"

	"github.com/yetone/magpie/internal/settings"
)

// Codex's lead picks the model of each subagent it spawns by itself, in
// spawn_agent's model argument (willz on Discord: gpt-6-astra), and its
// own config can't overrule that pick: [agents] default_subagent_model is
// only the default for a spawn that names none. settings.CodexSubagentModel
// is the model magpie puts Codex's subagents on, whatever the lead asked
// for: the gateway rewrites a subagent's request (x-openai-subagent
// collab_spawn) to it. A subagent's task is sealed by the ChatGPT backend
// when its lead ran there, and only a ChatGPT account opens it again, so
// the model is one a ChatGPT account in magpie serves (CodexSubagentModels).
// "" leaves every subagent on the model its lead asked for.

// CodexSubagentModel is the model Codex's subagents are put on, "" for
// the one the lead asks for.
func CodexSubagentModel() string {
	return strings.TrimSpace(heldSettings().CodexSubagentModel)
}

// CodexSubagentModels are the models a ChatGPT account in magpie serves,
// the ones Codex's subagents may be put on, in the catalog's order.
func CodexSubagentModels() []Entry {
	var out []Entry
	for _, e := range Served() {
		if e.Group == "" && e.Provider.Account != nil && e.Provider.Account.Agent == "codex" {
			out = append(out, e)
		}
	}
	return out
}

// CodexSubagentUnserved says why Codex's subagents can't be put on id now,
// "" when a ChatGPT account in magpie serves it.
func CodexSubagentUnserved(id string) string {
	for _, e := range CodexSubagentModels() {
		if e.ID == id {
			return ""
		}
	}
	if len(CodexSubagentModels()) == 0 {
		return "no ChatGPT account is on in magpie"
	}
	return "no ChatGPT account on in magpie serves it"
}

// SetCodexSubagentModel puts Codex's subagents on id, a ChatGPT account's
// model; "" leaves them on the model their lead asks for. One no ChatGPT
// account in magpie serves now is kept all the same — a backup or profile
// put back on a computer whose accounts differ, or an account signed in
// again later — and the gateway leaves subagents on the model asked for
// until one serves it, saying why in the route trace. Another provider's
// model is refused: it couldn't read a subagent's sealed task.
func SetCodexSubagentModel(id string) error {
	id = strings.TrimSpace(id)
	if id != "" && !codexAccountModel(id) {
		return fmt.Errorf("Codex's subagents can't be put on %s: a subagent's task is sealed for ChatGPT accounts, so its model must be one a ChatGPT account in magpie serves (codex/…)", id)
	}
	s := settings.Load()
	if s.CodexSubagentModel == id {
		return nil
	}
	s.CodexSubagentModel = id
	if err := settings.Save(s); err != nil {
		return err
	}
	Changed() // a held settings read sees it
	return nil
}

// codexAccountModel says id names a model of a ChatGPT account's, served
// now or not: codex/<model>, or one of a provider that is a ChatGPT
// account.
func codexAccountModel(id string) bool {
	pid, model, ok := strings.Cut(id, "/")
	if !ok || model == "" {
		return false
	}
	if pid == "codex" {
		return true
	}
	p, err := Find(pid)
	return err == nil && p.Account != nil && p.Account.Agent == "codex"
}
