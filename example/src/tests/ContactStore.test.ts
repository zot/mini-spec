// ContactStore Tests
// Test: test-contacts.md

import { ContactStore } from '../stores/ContactStore.js';

// Mock localStorage
const mockStorage: Record<string, string> = {};
const localStorageMock = {
  getItem: (key: string) => mockStorage[key] ?? null,
  setItem: (key: string, value: string) => { mockStorage[key] = value; },
  removeItem: (key: string) => { delete mockStorage[key]; },
  clear: () => { Object.keys(mockStorage).forEach(k => delete mockStorage[k]); }
};
Object.defineProperty(globalThis, 'localStorage', { value: localStorageMock });

// Mock crypto.randomUUID
let uuidCounter = 0;
Object.defineProperty(globalThis, 'crypto', {
  value: { randomUUID: () => `test-uuid-${++uuidCounter}` }
});

function createStore(): ContactStore {
  localStorageMock.clear();
  uuidCounter = 0;
  return new ContactStore();
}

// Test: test-contacts.md - CRUD Operations
console.log('=== ContactStore Tests ===\n');

// Test: add() creates contact with unique id
{
  const store = createStore();
  const c1 = store.add({ name: 'John', email: 'john@test.com', status: 'active', vip: false });
  const c2 = store.add({ name: 'Jane', email: 'jane@test.com', status: 'active', vip: true });
  console.assert(c1.id !== c2.id, 'add() creates unique ids');
  console.log('✓ add() creates contact with unique id');
}

// Test: add() persists to localStorage
{
  const store = createStore();
  store.add({ name: 'Test', email: 'test@test.com', status: 'active', vip: false });
  const saved = JSON.parse(mockStorage['contacts']);
  console.assert(saved.length === 1, 'add() persists to localStorage');
  console.log('✓ add() persists to localStorage');
}

// Test: getById() returns correct contact
{
  const store = createStore();
  const added = store.add({ name: 'Find Me', email: 'find@test.com', status: 'active', vip: false });
  const found = store.getById(added.id);
  console.assert(found?.name === 'Find Me', 'getById() returns correct contact');
  console.log('✓ getById() returns correct contact');
}

// Test: getById() returns undefined for missing id
{
  const store = createStore();
  const found = store.getById('nonexistent');
  console.assert(found === undefined, 'getById() returns undefined for missing');
  console.log('✓ getById() returns undefined for missing id');
}

// Test: update() modifies contact fields
{
  const store = createStore();
  const added = store.add({ name: 'Original', email: 'orig@test.com', status: 'active', vip: false });
  store.update(added.id, { name: 'Updated', vip: true });
  const updated = store.getById(added.id);
  console.assert(updated?.name === 'Updated' && updated?.vip === true, 'update() modifies fields');
  console.log('✓ update() modifies contact fields');
}

// Test: update() persists changes
{
  const store = createStore();
  const added = store.add({ name: 'Persist', email: 'p@test.com', status: 'active', vip: false });
  store.update(added.id, { status: 'inactive' });
  const saved = JSON.parse(mockStorage['contacts']);
  console.assert(saved[0].status === 'inactive', 'update() persists changes');
  console.log('✓ update() persists changes');
}

// Test: delete() removes contact
{
  const store = createStore();
  const added = store.add({ name: 'Delete Me', email: 'd@test.com', status: 'active', vip: false });
  store.delete(added.id);
  console.assert(store.getById(added.id) === undefined, 'delete() removes contact');
  console.log('✓ delete() removes contact');
}

// Test: delete() persists removal
{
  const store = createStore();
  const added = store.add({ name: 'Gone', email: 'g@test.com', status: 'active', vip: false });
  store.delete(added.id);
  const saved = JSON.parse(mockStorage['contacts']);
  console.assert(saved.length === 0, 'delete() persists removal');
  console.log('✓ delete() persists removal');
}

// Test: test-contacts.md - Filtering
console.log('\n--- Filtering ---');

// Test: getFiltered("") returns all contacts
{
  const store = createStore();
  store.add({ name: 'A', email: 'a@test.com', status: 'active', vip: false });
  store.add({ name: 'B', email: 'b@test.com', status: 'active', vip: false });
  console.assert(store.getFiltered('').length === 2, 'getFiltered("") returns all');
  console.log('✓ getFiltered("") returns all contacts');
}

// Test: getFiltered() matches name (case-insensitive)
{
  const store = createStore();
  store.add({ name: 'John Doe', email: 'john@test.com', status: 'active', vip: false });
  store.add({ name: 'Jane Smith', email: 'jane@test.com', status: 'active', vip: false });
  const results = store.getFiltered('john');
  console.assert(results.length === 1 && results[0].name === 'John Doe', 'filters by name');
  console.log('✓ getFiltered() matches name (case-insensitive)');
}

// Test: getFiltered() matches email (case-insensitive)
{
  const store = createStore();
  store.add({ name: 'User', email: 'HELLO@TEST.COM', status: 'active', vip: false });
  const results = store.getFiltered('hello');
  console.assert(results.length === 1, 'filters by email');
  console.log('✓ getFiltered() matches email (case-insensitive)');
}

// Test: getFiltered() returns empty for no matches
{
  const store = createStore();
  store.add({ name: 'Alice', email: 'alice@test.com', status: 'active', vip: false });
  console.assert(store.getFiltered('xyz').length === 0, 'returns empty for no matches');
  console.log('✓ getFiltered() returns empty for no matches');
}

// Test: test-contacts.md - Observers
console.log('\n--- Observers ---');

// Test: subscribe() callback called on add
{
  const store = createStore();
  let called = false;
  store.subscribe(() => { called = true; });
  store.add({ name: 'New', email: 'new@test.com', status: 'active', vip: false });
  console.assert(called, 'subscribe callback called on add');
  console.log('✓ subscribe() callback called on add');
}

// Test: subscribe() callback called on update
{
  const store = createStore();
  const added = store.add({ name: 'Up', email: 'up@test.com', status: 'active', vip: false });
  let called = false;
  store.subscribe(() => { called = true; });
  store.update(added.id, { name: 'Updated' });
  console.assert(called, 'subscribe callback called on update');
  console.log('✓ subscribe() callback called on update');
}

// Test: subscribe() callback called on delete
{
  const store = createStore();
  const added = store.add({ name: 'Del', email: 'del@test.com', status: 'active', vip: false });
  let called = false;
  store.subscribe(() => { called = true; });
  store.delete(added.id);
  console.assert(called, 'subscribe callback called on delete');
  console.log('✓ subscribe() callback called on delete');
}

// Test: unsubscribe prevents further callbacks
{
  const store = createStore();
  let count = 0;
  const unsub = store.subscribe(() => { count++; });
  store.add({ name: 'One', email: 'one@test.com', status: 'active', vip: false });
  unsub();
  store.add({ name: 'Two', email: 'two@test.com', status: 'active', vip: false });
  console.assert(count === 1, 'unsubscribe prevents callbacks');
  console.log('✓ unsubscribe prevents further callbacks');
}

console.log('\n=== All tests passed ===');
