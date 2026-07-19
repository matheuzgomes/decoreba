#!/usr/bin/env node
const fs = require('fs')
const path = require('path')
const https = require('https')
const { createGunzip } = require('zlib')
const { spawnSync } = require('child_process')

const pkg = require('../package.json')

const ASSETS = {
  'linux:x64':   { name: 'decoreba-linux-amd64',   ext: '.tar.gz' },
  'linux:arm64': { name: 'decoreba-linux-arm64',   ext: '.tar.gz' },
  'darwin:x64':  { name: 'decoreba-darwin-amd64',  ext: '.tar.gz' },
  'darwin:arm64':{ name: 'decoreba-darwin-arm64',  ext: '.tar.gz' },
  'win32:x64':   { name: 'decoreba-windows-amd64', ext: '.zip' },
}

function die(msg) {
  console.error(msg)
  process.exit(1)
}

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest)
    https.get(url, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        file.close()
        fs.unlinkSync(dest)
        return download(res.headers.location, dest).then(resolve).catch(reject)
      }
      if (res.statusCode !== 200) {
        file.close()
        fs.unlinkSync(dest)
        return reject(new Error(`HTTP ${res.statusCode}`))
      }
      res.pipe(file)
      file.on('finish', () => { file.close(); resolve() })
    }).on('error', (err) => {
      file.close()
      fs.unlinkSync(dest, () => {})
      reject(err)
    })
  })
}

function extractTarGz(archive, outputDir) {
  const tar = spawnSync('tar', ['xzf', archive, '-C', outputDir], { stdio: 'pipe' })
  if (tar.status !== 0) {
    die(`tar extraction failed: ${tar.stderr?.toString() || tar.stdout?.toString()}`)
  }
}

function extractZip(archive, outputDir) {
  let ok = false
  const powershell = spawnSync('powershell', [
    '-NoProfile', '-NonInteractive',
    '-Command',
    `Expand-Archive -Path "${archive}" -DestinationPath "${outputDir}" -Force`
  ], { stdio: 'pipe' })
  if (powershell.status === 0) ok = true

  if (!ok) {
    const compat = spawnSync('unzip', ['-o', archive, '-d', outputDir], { stdio: 'pipe' })
    if (compat.status === 0) ok = true
  }

  if (!ok) {
    die('failed to extract zip archive (tried both PowerShell Expand-Archive and unzip)')
  }
}

async function main() {
  const key = `${process.platform}:${process.arch}`
  const info = ASSETS[key]
  if (!info) {
    die(`unsupported platform: ${process.platform} ${process.arch}`)
  }

  const version = pkg.version
  const binDir = path.join(__dirname, '..', 'bin')
  const exeName = process.platform === 'win32' ? 'decoreba.exe' : 'decoreba'
  const binPath = path.join(binDir, exeName)

  if (fs.existsSync(binPath)) {
    console.log(`decoreba v${version} already installed`)
    return
  }

  const ext = info.ext
  const archiveName = `${info.name}${ext}`
  const url = `https://github.com/matheuzgomes/decoreba/releases/download/v${version}/${archiveName}`

  console.log(`downloading decoreba v${version} for ${process.platform}/${process.arch}...`)

  const tmpDir = fs.mkdtempSync(path.join(__dirname, '.decoreba-install-'))
  const archivePath = path.join(tmpDir, archiveName)

  try {
    await download(url, archivePath)
    console.log('extracting...')

    if (ext === '.zip') {
      extractZip(archivePath, tmpDir)
    } else {
      extractTarGz(archivePath, tmpDir)
    }

    const extracted = path.join(tmpDir, info.name)
    if (!fs.existsSync(extracted)) {
      die(`extracted binary not found at ${extracted}`)
    }

    fs.renameSync(extracted, binPath)
    fs.chmodSync(binPath, 0o755)
    console.log(`installed to ${binPath}`)
  } catch (err) {
    die(`download failed: ${err.message}`)
  } finally {
    fs.rmSync(tmpDir, { recursive: true, force: true })
  }
}

main()
