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

/** Most nudges or cheers shown at once; older ones are dropped as seen. */
const MAX_SHOWN = 4;

/**
 * Splits items (any order, numeric id) into the newest `max` to show, oldest
 * first, and the older ones to drop.
 */
function newestFirstSplit(items, max = MAX_SHOWN) {
  const byId = [...items].sort((a, b) => a.id - b.id);
  const cut = Math.max(0, byId.length - max);
  return { shown: byId.slice(cut), stale: byId.slice(0, cut) };
}

module.exports = { lastClaudeEvent, cardStep, newestFirstSplit, MAX_SHOWN };
