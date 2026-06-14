// generate-icons.js - 从 SVG 生成多尺寸 PNG 图标
import sharp from 'sharp'
import { readFileSync } from 'fs'
import { resolve, dirname } from 'path'
import { fileURLToPath } from 'url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const svgPath = resolve(__dirname, 'public/icon.svg')
const svgBuffer = readFileSync(svgPath)

const sizes = [
  { name: 'icon-192.png', size: 192 },
  { name: 'icon-512.png', size: 512 },
  { name: 'apple-touch-icon.png', size: 180 },
  { name: 'favicon-32.png', size: 32 },
  { name: 'favicon-16.png', size: 16 },
]

for (const { name, size } of sizes) {
  await sharp(svgBuffer, { density: 300 })
    .resize(size, size)
    .png()
    .toFile(resolve(__dirname, 'public', name))
  console.log(`Generated: public/${name} (${size}x${size})`)
}

console.log('Done!')
