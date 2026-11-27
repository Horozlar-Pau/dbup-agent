'use strict';

const test = require('node:test');
const assert = require('node:assert');
const { cardStep } = require('./watch');

const ev = (type, at) => ({ type, at, source: type.startsWith('git') ? 'git' : 'claude' });
const q = [{ id: 1 }];

test('card opens while Claude works and nudges wait', () => {
  const s = {};
  assert.equal(cardStep(s, [ev('ai.prompt_started', 't1')], q, false), 'show');
  assert.equal(cardStep(s, [ev('ai.prompt_started', 't1')], [], false), 'none');
});

test('card closes when Claude finishes', () => {
  const s = {};
  assert.equal(cardStep(s, [ev('ai.prompt_started', 't1'), ev('ai.prompt_finished', 't2')], q, true), 'close');
});

test('git commits do not change whether Claude works', () => {
  const s = {};
  assert.equal(cardStep(s, [ev('ai.prompt_started', 't1'), ev('git.commit', 't2')], q, false), 'show');
});

test('a dismissed card stays closed until the next prompt', () => {
  const s = {};
  cardStep(s, [ev('ai.prompt_started', 't1')], q, false);
  s.dismissed = true;
  assert.equal(cardStep(s, [ev('ai.prompt_started', 't1')], q, false), 'none');
  assert.equal(cardStep(s, [ev('ai.prompt_started', 't3')], q, false), 'show');
});
