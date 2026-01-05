interface Contact {
  name: string;
  email: string;
  status: "active" | "inactive";
  vip: boolean;
}

const contacts: Contact[] = [];
let editing: Contact | null = null;
let current: Contact | null = null;

const $ = <T extends HTMLElement>(id: string) => document.getElementById(id) as T;

const searchInput = $<HTMLInputElement>("search");
const countBadge = $("count");
const listEl = $("list");
const detailEl = $("detail");
const nameInput = $<HTMLInputElement>("name");
const emailInput = $<HTMLInputElement>("email");
const statusSelect = $<HTMLSelectElement>("status");
const vipCheckbox = $<HTMLInputElement>("vip");
const darkCheckbox = $<HTMLInputElement>("darkMode");

function filteredContacts(): Contact[] {
  const q = searchInput.value.toLowerCase();
  if (!q) return contacts;
  return contacts.filter(c =>
    c.name.toLowerCase().includes(q) || c.email.toLowerCase().includes(q)
  );
}

function render() {
  const filtered = filteredContacts();
  countBadge.textContent = String(filtered.length);

  listEl.innerHTML = filtered.map(c => `
    <div class="contact-row ${c === editing ? "selected" : ""}" data-index="${contacts.indexOf(c)}">
      <div class="contact-name">${esc(c.name)}</div>
      <div class="contact-email">${esc(c.email)}</div>
    </div>
  `).join("");

  listEl.querySelectorAll(".contact-row").forEach(row => {
    row.addEventListener("click", () => select(contacts[Number((row as HTMLElement).dataset.index)]));
  });
}

function esc(s: string): string {
  return s.replace(/[&<>"']/g, c => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c] || c));
}

function showDetail() {
  detailEl.classList.remove("hidden");
  if (current) {
    nameInput.value = current.name;
    emailInput.value = current.email;
    statusSelect.value = current.status;
    vipCheckbox.checked = current.vip;
  }
}

function hideDetail() {
  detailEl.classList.add("hidden");
  current = null;
  editing = null;
  render();
}

function select(contact: Contact) {
  editing = contact;
  current = { ...contact };
  showDetail();
  render();
}

function add() {
  editing = null;
  current = { name: "New Contact", email: "", status: "active", vip: false };
  showDetail();
}

function save() {
  if (!current) return;
  current.name = nameInput.value;
  current.email = emailInput.value;
  current.status = statusSelect.value as "active" | "inactive";
  current.vip = vipCheckbox.checked;

  if (editing) {
    Object.assign(editing, current);
  } else {
    contacts.push(current);
  }
  hideDetail();
}

function deleteCurrent() {
  if (editing) {
    const idx = contacts.indexOf(editing);
    if (idx >= 0) contacts.splice(idx, 1);
  }
  hideDetail();
}

// Event bindings
searchInput.addEventListener("input", render);
$("addBtn").addEventListener("click", add);
$("saveBtn").addEventListener("click", save);
$("cancelBtn").addEventListener("click", hideDetail);
$("deleteBtn").addEventListener("click", deleteCurrent);
darkCheckbox.addEventListener("change", () => document.body.classList.toggle("dark", darkCheckbox.checked));

render();
