#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

const BINARY_NAME = 'cortex';

function getBinaryPath() {
  const binDir = path.join(__dirname, '..', 'bin');
  const ext = process.platform === 'win32' ? '.exe' : '';
  return path.join(binDir, BINARY_NAME + ext);
}

try {
  const binaryPath = getBinaryPath();
  if (fs.existsSync(binaryPath)) {
    fs.unlinkSync(binaryPath);
    console.log('Cortex uninstalled successfully');
  }
} catch (error) {
  // Ignore errors during uninstall
}
