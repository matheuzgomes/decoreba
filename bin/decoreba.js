#!/usr/bin/env node
const { spawnSync } = require('child_process')
const path = require('path')
const fs = require('fs')

const binName = process.platform === 'win32' ? 'decoreba.exe' : 'decoreba'
const binPath = path.join(__dirname, binName)

if (!fs.existsSync(binPath)) {
  console.error('decoreba binary not found. Run npm install to download it.')
  process.exit(1)
}

const { status } = spawnSync(binPath, process.argv.slice(2), { stdio: 'inherit' })
process.exit(status ?? 1)
