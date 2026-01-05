// DetailPanel - View component
// CRC: crc-DetailPanel.md | Seq: seq-crud.md

import { Contact } from '../models/Contact.js';
import { ContactStore } from '../stores/ContactStore.js';

export class DetailPanel {
  private currentId?: string;
  private isNew = false;
  private saveCallback?: (contact: Contact) => void;
  private deleteCallback?: (id: string) => void;
  private cancelCallback?: () => void;

  private nameInput!: HTMLInputElement;
  private emailInput!: HTMLInputElement;
  private statusActive!: HTMLInputElement;
  private statusInactive!: HTMLInputElement;
  private vipCheckbox!: HTMLInputElement;
  private deleteBtn!: HTMLButtonElement;

  // CRC: crc-DetailPanel.md
  constructor(
    private container: HTMLElement,
    private store: ContactStore
  ) {
    this.setupForm();
    this.hide();
  }

  // CRC: crc-DetailPanel.md
  private setupForm(): void {
    this.container.innerHTML = `
      <form id="detail-form">
        <div class="form-group">
          <label for="name">Name</label>
          <input type="text" id="name" name="name" required>
        </div>
        <div class="form-group">
          <label for="email">Email</label>
          <input type="email" id="email" name="email" required>
        </div>
        <div class="form-group">
          <label>Status</label>
          <div class="radio-group">
            <label><input type="radio" name="status" value="active" checked> Active</label>
            <label><input type="radio" name="status" value="inactive"> Inactive</label>
          </div>
        </div>
        <div class="form-group">
          <label><input type="checkbox" id="vip" name="vip"> VIP</label>
        </div>
        <div class="form-actions">
          <button type="submit" class="btn-primary">Save</button>
          <button type="button" class="btn-secondary" id="cancel-btn">Cancel</button>
          <button type="button" class="btn-danger" id="delete-btn">Delete</button>
        </div>
      </form>
    `;

    this.nameInput = this.container.querySelector('#name')!;
    this.emailInput = this.container.querySelector('#email')!;
    this.statusActive = this.container.querySelector('input[value="active"]')!;
    this.statusInactive = this.container.querySelector('input[value="inactive"]')!;
    this.vipCheckbox = this.container.querySelector('#vip')!;
    this.deleteBtn = this.container.querySelector('#delete-btn')!;

    const form = this.container.querySelector('form')!;
    form.addEventListener('submit', (e: Event) => {
      e.preventDefault();
      this.handleSave();
    });

    this.container.querySelector('#cancel-btn')!.addEventListener('click', () => {
      this.cancelCallback?.();
    });

    this.deleteBtn.addEventListener('click', () => {
      if (this.currentId) {
        this.deleteCallback?.(this.currentId);
      }
    });
  }

  // CRC: crc-DetailPanel.md | Seq: seq-crud.md
  show(contact: Contact | null): void {
    this.container.classList.remove('hidden');
    this.isNew = contact === null;
    this.currentId = contact?.id;

    if (contact) {
      this.nameInput.value = contact.name;
      this.emailInput.value = contact.email;
      if (contact.status === 'active') {
        this.statusActive.checked = true;
      } else {
        this.statusInactive.checked = true;
      }
      this.vipCheckbox.checked = contact.vip;
      this.deleteBtn.classList.remove('hidden');
    } else {
      this.nameInput.value = '';
      this.emailInput.value = '';
      this.statusActive.checked = true;
      this.vipCheckbox.checked = false;
      this.deleteBtn.classList.add('hidden');
    }
    this.nameInput.focus();
  }

  // CRC: crc-DetailPanel.md
  hide(): void {
    this.container.classList.add('hidden');
    this.currentId = undefined;
    this.isNew = false;
  }

  // CRC: crc-DetailPanel.md
  onSave(callback: (contact: Contact) => void): void {
    this.saveCallback = callback;
  }

  // CRC: crc-DetailPanel.md
  onDelete(callback: (id: string) => void): void {
    this.deleteCallback = callback;
  }

  // CRC: crc-DetailPanel.md
  onCancel(callback: () => void): void {
    this.cancelCallback = callback;
  }

  // CRC: crc-DetailPanel.md | Seq: seq-crud.md
  private handleSave(): void {
    const data = {
      name: this.nameInput.value.trim(),
      email: this.emailInput.value.trim(),
      status: (this.statusActive.checked ? 'active' : 'inactive') as 'active' | 'inactive',
      vip: this.vipCheckbox.checked
    };

    if (this.isNew) {
      const contact = this.store.add(data);
      this.saveCallback?.(contact);
    } else if (this.currentId) {
      this.store.update(this.currentId, data);
      const updated = this.store.getById(this.currentId);
      if (updated) this.saveCallback?.(updated);
    }
  }
}
