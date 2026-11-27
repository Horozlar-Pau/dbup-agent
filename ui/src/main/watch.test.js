'use strict';

const test = require('node:test');
const assert = require('node:assert');
const { cardStep, lastClaudeEvent } = require('./watch');

test('card opens as soon as a nudge arrives', () => {
  const s = {};
  assert.equal(cardStep(s, [{ id: 1 }], false), 'show');
});

test('card closes when nothing is waiting', () => {
  const s = {};
  assert.equal(cardStep(s, [], true), 'close');
  assert.equal(cardStep(s, [], false), 'none');
});

test('a dismissed card stays closed until a new nudge arrives', () => {
  const s = {};
  cardStep(s, [{ id: 1 }], false);
  s.dismissed = true;
  assert.equal(cardStep(s, [{ id: 1 }], false), 'none');
  assert.equal(cardStep(s, [{ id: 1 }, { id: 2 }], false), 'show');
});

test('lastClaudeEvent skips git events', () => {
  const ev = [{ type: 'ai.prompt_started', at: 't1' }, { type: 'git.commit', at: 't2' }];
  assert.equal(lastClaudeEvent(ev).type, 'ai.prompt_started');
});

test('only the newest four are shown, older ones are dropped', () => {
  const { newestFirstSplit } = require('./watch');
  const items = [5, 1, 3, 2, 4, 6].map((id) => ({ id }));
  const { shown, stale } = newestFirstSplit(items);
  assert.deepEqual(shown.map((n) => n.id), [3, 4, 5, 6]);
  assert.deepEqual(stale.map((n) => n.id), [1, 2]);
});

test('nothing is dropped when four or fewer are waiting', () => {
  const { newestFirstSplit } = require('./watch');
  assert.equal(newestFirstSplit([{ id: 1 }, { id: 2 }]).stale.length, 0);
});
