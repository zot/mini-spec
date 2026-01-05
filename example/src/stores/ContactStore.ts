// ContactStore - Observable data store
// CRC: crc-ContactStore.md | Seq: seq-crud.md, seq-search.md

import { Contact, createContact } from '../models/Contact.js';

const STORAGE_KEY = 'contacts';

export class ContactStore {
  private contacts: Contact[] = [];
  private observers: Set<() => void> = new Set();

  constructor() {
    this.load();
  }

  // CRC: crc-ContactStore.md
  subscribe(callback: () => void): () => void {
    this.observers.add(callback);
    return () => this.observers.delete(callback);
  }

  // CRC: crc-ContactStore.md
  private notify(): void {
    this.observers.forEach(cb => cb());
  }

  // CRC: crc-ContactStore.md
  private load(): void {
    const data = localStorage.getItem(STORAGE_KEY);
    this.contacts = data ? JSON.parse(data) : [];
  }

  // CRC: crc-ContactStore.md
  private persist(): void {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(this.contacts));
  }

  // CRC: crc-ContactStore.md | Seq: seq-crud.md
  getAll(): Contact[] {
    return [...this.contacts];
  }

  // CRC: crc-ContactStore.md | Seq: seq-search.md
  getFiltered(term: string): Contact[] {
    if (!term) return this.getAll();
    const lower = term.toLowerCase();
    return this.contacts.filter(c =>
      c.name.toLowerCase().includes(lower) ||
      c.email.toLowerCase().includes(lower)
    );
  }

  // CRC: crc-ContactStore.md | Seq: seq-crud.md
  getById(id: string): Contact | undefined {
    return this.contacts.find(c => c.id === id);
  }

  // CRC: crc-ContactStore.md | Seq: seq-crud.md
  add(data: Omit<Contact, 'id'>): Contact {
    const contact = createContact(data);
    this.contacts.push(contact);
    this.persist();
    this.notify();
    return contact;
  }

  // CRC: crc-ContactStore.md | Seq: seq-crud.md
  update(id: string, data: Partial<Omit<Contact, 'id'>>): void {
    const index = this.contacts.findIndex(c => c.id === id);
    if (index !== -1) {
      this.contacts[index] = { ...this.contacts[index], ...data };
      this.persist();
      this.notify();
    }
  }

  // CRC: crc-ContactStore.md | Seq: seq-crud.md
  delete(id: string): void {
    this.contacts = this.contacts.filter(c => c.id !== id);
    this.persist();
    this.notify();
  }
}
