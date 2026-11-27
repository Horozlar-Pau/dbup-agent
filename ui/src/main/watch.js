// Pure decisions for the overlay windows, kept apart from Electron so they can be tested.
'use strict';

const CLAUDE_TYPES = new Set(['ai.prompt_started', 'ai.prompt_finished', 'ai.session_ended']);

// Latest Claude event in the list (oldest first), or null.
function lastClaudeEvent(events) {
  for (let i = events.length - 1; i >= 0; i--) if (CLAUDE_TYPES.has(events[i].type)) return events[i];
  return null;
}

/**
 * Decides what the wait card should do. Nudges are shown as soon as they
 * arrive, whether or not Claude is working (DBUP-44).
 * state: { lastId, dismissed } carried between polls (mutated).
 * Returns 'show' | 'close' | 'none'.
 */
function cardStep(state, queued, cardOpen) {
  if (queued.length === 0) return cardOpen ? 'close' : 'none';
  const newest = Math.max(...queued.map((n) => n.id));
  if (newest !== state.lastId) {
    state.lastId = newest; // a new nudge: a closed card opens again
    state.dismissed = false;
  }
  return state.dismissed ? 'none' : 'show';
}

module.exports = { lastClaudeEvent, cardStep };
