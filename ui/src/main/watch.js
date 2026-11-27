// Pure decisions for the overlay windows, kept apart from Electron so they can be tested.
'use strict';

const CLAUDE_TYPES = new Set(['ai.prompt_started', 'ai.prompt_finished', 'ai.session_ended']);

// Latest Claude event in the list (oldest first), or null.
function lastClaudeEvent(events) {
  for (let i = events.length - 1; i >= 0; i--) if (CLAUDE_TYPES.has(events[i].type)) return events[i];
  return null;
}

/**
 * Decides what the wait card should do.
 * state: { promptAt, dismissed } carried between polls (mutated).
 * Returns 'show' | 'close' | 'none'.
 */
function cardStep(state, events, queued, cardOpen) {
  const last = lastClaudeEvent(events);
  const working = last?.type === 'ai.prompt_started';
  if (working && last.at !== state.promptAt) {
    state.promptAt = last.at; // a new prompt: a closed card may open again
    state.dismissed = false;
  }
  if (!working) return cardOpen ? 'close' : 'none';
  if (state.dismissed || queued.length === 0) return 'none';
  return 'show';
}

module.exports = { lastClaudeEvent, cardStep };
