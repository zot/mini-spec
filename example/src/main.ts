// Main entry point
// CRC: crc-App.md

import { App } from './App.js';

document.addEventListener('DOMContentLoaded', () => {
  const root = document.getElementById('app')!;
  const app = new App(root);
  app.init();
});
