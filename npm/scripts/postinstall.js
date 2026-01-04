#!/usr/bin/env node

const https = require('https');
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');
const zlib = require('zlib');

const REPO = 'adityaraj/cortex';
const BINARY_NAME = 'cortex';
const VERSION = require('../package.json').version;

// Platform mapping
const PLATFORM_MAP = {
  darwin: 'darwin',
  linux: 'linux',
  win32: 'windows',
};

const ARCH_MAP = {
  x64: 'amd64',
  arm64: 'arm64',
};

function getPlatform() {
  const platform = PLATFORM_MAP[process.platform];
  const arch = ARCH_MAP[process.arch];

  if (!platform) {
    throw new Error(`Unsupported platform: ${process.platform}`);
  }
  if (!arch) {
    throw new Error(`Unsupported architecture: ${process.arch}`);
  }

  return { platform, arch };
}

function getBinaryPath() {
  const binDir = path.join(__dirname, '..', 'bin');
  const ext = process.platform === 'win32' ? '.exe' : '';
  return path.join(binDir, BINARY_NAME + ext);
}

function downloadFile(url) {
  return new Promise((resolve, reject) => {
    const request = (url) => {
      https.get(url, { headers: { 'User-Agent': 'cortex-npm' } }, (res) => {
        // Handle redirects
        if (res.statusCode === 301 || res.statusCode === 302) {
          request(res.headers.location);
          return;
        }

        if (res.statusCode !== 200) {
          reject(new Error(`Failed to download: ${res.statusCode}`));
          return;
        }

        const chunks = [];
        res.on('data', (chunk) => chunks.push(chunk));
        res.on('end', () => resolve(Buffer.concat(chunks)));
        res.on('error', reject);
      }).on('error', reject);
    };
    request(url);
  });
}

async function extractTarGz(buffer, destPath) {
  const gunzip = zlib.createGunzip();
  const tempDir = path.join(__dirname, '..', '.temp');

  // Create temp directory
  if (!fs.existsSync(tempDir)) {
    fs.mkdirSync(tempDir, { recursive: true });
  }

  const tarPath = path.join(tempDir, 'archive.tar');

  // Decompress gzip
  await new Promise((resolve, reject) => {
    const input = require('stream').Readable.from(buffer);
    const output = fs.createWriteStream(tarPath);
    input.pipe(gunzip).pipe(output);
    output.on('finish', resolve);
    output.on('error', reject);
  });

  // Extract tar using system tar (more reliable than JS implementations)
  try {
    execSync(`tar -xf "${tarPath}" -C "${tempDir}"`, { stdio: 'pipe' });
  } catch (e) {
    throw new Error('Failed to extract archive. Make sure tar is installed.');
  }

  // Find and move the binary
  const files = fs.readdirSync(tempDir);
  const binaryFile = files.find(f => f.startsWith(BINARY_NAME) && !f.endsWith('.tar'));

  if (!binaryFile) {
    throw new Error('Binary not found in archive');
  }

  const srcPath = path.join(tempDir, binaryFile);
  fs.copyFileSync(srcPath, destPath);
  fs.chmodSync(destPath, 0o755);

  // Cleanup
  fs.rmSync(tempDir, { recursive: true, force: true });
}

async function extractZip(buffer, destPath) {
  const AdmZip = require('adm-zip');
  const zip = new AdmZip(buffer);
  const entries = zip.getEntries();

  const binaryEntry = entries.find(e => e.entryName.includes(BINARY_NAME));
  if (!binaryEntry) {
    throw new Error('Binary not found in archive');
  }

  const binDir = path.dirname(destPath);
  zip.extractEntryTo(binaryEntry, binDir, false, true);

  // Rename if needed
  const extractedPath = path.join(binDir, binaryEntry.entryName);
  if (extractedPath !== destPath) {
    fs.renameSync(extractedPath, destPath);
  }
}

async function install() {
  console.log('Installing Cortex...');

  const { platform, arch } = getPlatform();
  const binaryPath = getBinaryPath();
  const binDir = path.dirname(binaryPath);

  // Ensure bin directory exists
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }

  // Determine download URL
  const ext = platform === 'windows' ? 'zip' : 'tar.gz';
  const downloadUrl = `https://github.com/${REPO}/releases/download/v${VERSION}/${BINARY_NAME}-${platform}-${arch}.${ext}`;

  console.log(`Downloading from: ${downloadUrl}`);

  try {
    const buffer = await downloadFile(downloadUrl);
    console.log('Download complete. Extracting...');

    if (ext === 'zip') {
      await extractZip(buffer, binaryPath);
    } else {
      await extractTarGz(buffer, binaryPath);
    }

    console.log(`Cortex installed successfully!`);
    console.log(`Binary location: ${binaryPath}`);

    // Verify installation
    try {
      const version = execSync(`"${binaryPath}" --version`, { encoding: 'utf8' }).trim();
      console.log(`Version: ${version}`);
    } catch (e) {
      // Ignore verification errors
    }
  } catch (error) {
    console.error('Failed to install Cortex:', error.message);
    console.error('');
    console.error('You can install manually:');
    console.error('  curl -fsSL https://raw.githubusercontent.com/adityaraj/cortex/main/install.sh | bash');
    process.exit(1);
  }
}

install();
