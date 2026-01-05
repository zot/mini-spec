// App - Coordinator
// CRC: crc-App.md | Seq: seq-crud.md, seq-search.md

import { ContactStore } from './stores/ContactStore.js';
import { ContactListView } from './views/ContactListView.js';
import { DetailPanel } from './views/DetailPanel.js';

const DARK_MODE_KEY = 'darkMode';

export class App {
  private store: ContactStore;
  private listView: ContactListView;
  private detailPanel: DetailPanel;
  private searchInput!: HTMLInputElement;
  private badge!: HTMLElement;
  private darkToggle!: HTMLInputElement;

  // CRC: crc-App.md
  constructor(private root: HTMLElement) {
    this.store = new ContactStore();
    this.listView = new ContactListView(
      this.root.querySelector('#contact-list')!,
      this.store
    );
    this.detailPanel = new DetailPanel(
      this.root.querySelector('#detail-panel')!,
      this.store
    );
  }

  // CRC: crc-App.md
  init(): void {
    this.searchInput = this.root.querySelector('#search')!;
    this.badge = this.root.querySelector('#badge')!;
    this.darkToggle = this.root.querySelector('#dark-toggle')!;

    this.setupEventHandlers();
    this.loadDarkMode();
    this.updateBadge();
    this.listView.render();
  }

  // CRC: crc-App.md | Seq: seq-crud.md, seq-search.md
  private setupEventHandlers(): void {
    // Search
    this.searchInput.addEventListener('input', () => {
      this.listView.setSearchTerm(this.searchInput.value);
      this.updateBadge();
    });

    // Add button
    this.root.querySelector('#add-btn')!.addEventListener('click', () => {
      this.listView.setSelectedId(undefined);
      this.detailPanel.show(null);
    });

    // List selection
    this.listView.onSelect((id) => {
      const contact = this.store.getById(id);
      if (contact) {
        this.listView.setSelectedId(id);
        this.detailPanel.show(contact);
      }
    });

    // Detail panel events
    this.detailPanel.onSave(() => {
      this.detailPanel.hide();
      this.listView.setSelectedId(undefined);
      this.updateBadge();
    });

    this.detailPanel.onDelete((id) => {
      this.store.delete(id);
      this.detailPanel.hide();
      this.listView.setSelectedId(undefined);
      this.updateBadge();
    });

    this.detailPanel.onCancel(() => {
      this.detailPanel.hide();
      this.listView.setSelectedId(undefined);
    });

    // Dark mode
    this.darkToggle.addEventListener('change', () => {
      this.setDarkMode(this.darkToggle.checked);
    });

    // Store changes
    this.store.subscribe(() => {
      this.updateBadge();
    });
  }

  // CRC: crc-App.md
  private updateBadge(): void {
    this.badge.textContent = String(this.listView.getFilteredCount());
  }

  // CRC: crc-App.md
  private loadDarkMode(): void {
    const saved = localStorage.getItem(DARK_MODE_KEY);
    const isDark = saved === 'true';
    this.darkToggle.checked = isDark;
    this.setDarkMode(isDark);
  }

  // CRC: crc-App.md
  private setDarkMode(enabled: boolean): void {
    document.body.classList.toggle('dark', enabled);
    localStorage.setItem(DARK_MODE_KEY, String(enabled));
  }
}
