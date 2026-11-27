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
