// Contact - Entity model
// CRC: crc-Contact.md

export interface Contact {
  id: string;
  name: string;
  email: string;
  status: 'active' | 'inactive';
  vip: boolean;
}

// CRC: crc-Contact.md
export function createContact(data: Omit<Contact, 'id'>): Contact {
  return {
    id: crypto.randomUUID(),
    ...data
  };
}
