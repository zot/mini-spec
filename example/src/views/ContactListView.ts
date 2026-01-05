// ContactListView - View component
// CRC: crc-ContactListView.md | Seq: seq-crud.md, seq-search.md

import { Contact } from '../models/Contact.js';
import { ContactStore } from '../stores/ContactStore.js';

export class ContactListView {
  private searchTerm = '';
  private selectCallback?: (id: string) => void;
  private selectedId?: string;

  // CRC: crc-ContactListView.md
  constructor(
    private container: HTMLElement,
    private store: ContactStore
  ) {
    this.store.subscribe(() => this.render());
  }

  // CRC: crc-ContactListView.md | Seq: seq-search.md
  setSearchTerm(term: string): void {
    this.searchTerm = term;
    this.render();
  }

  // CRC: crc-ContactListView.md
  setSelectedId(id: string | undefined): void {
    this.selectedId = id;
    this.render();
  }

  // CRC: crc-ContactListView.md
  onSelect(callback: (id: string) => void): void {
    this.selectCallback = callback;
  }

  // CRC: crc-ContactListView.md
  getFilteredCount(): number {
    return this.store.getFiltered(this.searchTerm).length;
  }

  // CRC: crc-ContactListView.md | Seq: seq-crud.md
  render(): void {
    const contacts = this.store.getFiltered(this.searchTerm);
    this.container.innerHTML = '';

    contacts.forEach(contact => {
      const item = this.createListItem(contact);
      this.container.appendChild(item);
    });
  }

  // CRC: crc-ContactListView.md
  private createListItem(contact: Contact): HTMLElement {
    const item = document.createElement('div');
    item.className = 'contact-item' + (contact.id === this.selectedId ? ' selected' : '');
    item.innerHTML = `
      <div class="contact-name${contact.vip ? ' vip' : ''}">${this.escape(contact.name)}</div>
      <div class="contact-email">${this.escape(contact.email)}</div>
    `;
    item.addEventListener('click', () => {
      this.selectCallback?.(contact.id);
    });
    return item;
  }

  // CRC: crc-ContactListView.md
  private escape(str: string): string {
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
  }
}
