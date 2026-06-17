import katex from 'katex'
import 'katex/dist/katex.min.css'

/**
 * 渲染文本中的 Markdown 图片 ![alt](url)
 * 输出带样式的 <img> 标签，适配暗色模式
 * @param {string} text - 原始文本
 * @returns {string} 含 <img> 标签的文本
 */
export function renderImages(text) {
  if (!text) return ''
  return text.replace(/!\[([^\]]*)\]\(([^)]+)\)/g, (_, alt, url) => {
    const escapedAlt = alt.replace(/"/g, '&quot;').replace(/</g, '&lt;')
    const escapedUrl = url.replace(/"/g, '&quot;').replace(/</g, '&lt;')
    return `<img src="${escapedUrl}" alt="${escapedAlt}" class="card-image" loading="lazy" />`
  })
}

/**
 * 渲染文本中的 LaTeX 数学公式 + Markdown 图片
 * 支持 $...$ 行内、$$...$$ 块级、\(...\) 行内、\[...\] 块级
 * 支持 ![alt](url) 图片语法
 * @param {string} text - 原始文本
 * @returns {string} 含 KaTeX HTML 和 <img> 的文本
 */
export function renderMath(text) {
  if (!text) return ''

  // 收集所有需保护的内容（数学公式+图片），用占位符替换
  const blocks = []
  let result = text

  // 0. Markdown 图片（先处理，避免 URL 中的 $ 等字符干扰数学公式匹配）
  result = result.replace(/!\[([^\]]*)\]\(([^)]+)\)/g, (_, alt, url) => {
    const idx = blocks.length
    const escapedAlt = alt.replace(/"/g, '&quot;').replace(/</g, '&lt;')
    const escapedUrl = url.replace(/"/g, '&quot;').replace(/</g, '&lt;')
    blocks.push(`<img src="${escapedUrl}" alt="${escapedAlt}" class="card-image" loading="lazy" />`)
    return `\x00BLOCK${idx}\x00`
  })

  // 1. 块级 $$...$$ 和 \[...\]
  result = result.replace(/\$\$([\s\S]*?)\$\$/g, (_, formula) => {
    const idx = blocks.length
    try {
      blocks.push(katex.renderToString(formula.trim(), { displayMode: true, throwOnError: false }))
    } catch {
      blocks.push(formula)
    }
    return `\x00BLOCK${idx}\x00`
  })
  result = result.replace(/\\\[([\s\S]*?)\\\]/g, (_, formula) => {
    const idx = blocks.length
    try {
      blocks.push(katex.renderToString(formula.trim(), { displayMode: true, throwOnError: false }))
    } catch {
      blocks.push(formula)
    }
    return `\x00BLOCK${idx}\x00`
  })

  // 2. 行内 $...$ 和 \(...\)
  result = result.replace(/\$([^$]+?)\$/g, (_, formula) => {
    const idx = blocks.length
    try {
      blocks.push(katex.renderToString(formula.trim(), { displayMode: false, throwOnError: false }))
    } catch {
      blocks.push(formula)
    }
    return `\x00BLOCK${idx}\x00`
  })
  result = result.replace(/\\\((.+?)\\\)/g, (_, formula) => {
    const idx = blocks.length
    try {
      blocks.push(katex.renderToString(formula.trim(), { displayMode: false, throwOnError: false }))
    } catch {
      blocks.push(formula)
    }
    return `\x00BLOCK${idx}\x00`
  })

  // 3. 还原占位符
  // eslint-disable-next-line no-control-regex
  result = result.replace(/\x00BLOCK(\d+)\x00/g, (_, idx) => blocks[parseInt(idx)] || '')

  return result
}

/**
 * 先提取数学公式和图片 → 对非数学/非图片部分执行 HTML 转义+搜索高亮 → 再还原
 * 用于 Cards.vue 的 highlight() 场景（搜索高亮 + LaTeX + 图片共存）
 * @param {string} text - 原始文本
 * @param {function} escapeAndHighlight - 对纯文本做转义+高亮的函数 (text) => html
 * @returns {string}
 */
export function renderMathWithHighlight(text, escapeAndHighlight) {
  if (!text) return ''

  // 拆分：交替的 [普通文本, 数学公式/图片, 普通文本, ...]
  const parts = []
  const regex = /(!\[[^\]]*\]\([^)]+\)|\$\$[\s\S]*?\$\$|\\\[[\s\S]*?\\\]|\$[^$]+?\$|\\\(.+?\\\))/g
  let lastIndex = 0
  let match

  while ((match = regex.exec(text)) !== null) {
    if (match.index > lastIndex) {
      parts.push({ type: 'text', content: text.slice(lastIndex, match.index) })
    }
    const content = match[0]
    if (content.startsWith('![')) {
      parts.push({ type: 'image', content })
    } else {
      parts.push({ type: 'math', content })
    }
    lastIndex = regex.lastIndex
  }
  if (lastIndex < text.length) {
    parts.push({ type: 'text', content: text.slice(lastIndex) })
  }

  // 分别处理
  return parts
    .map((part) => {
      if (part.type === 'text') {
        return escapeAndHighlight(part.content)
      }
      if (part.type === 'image') {
        return renderImages(part.content)
      }
      return renderMath(part.content)
    })
    .join('')
}
