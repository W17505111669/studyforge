import katex from 'katex'
import 'katex/dist/katex.min.css'

/**
 * 渲染文本中的 LaTeX 数学公式
 * 支持 $...$ 行内、$$...$$ 块级、\(...\) 行内、\[...\] 块级
 * @param {string} text - 原始文本
 * @returns {string} 含 KaTeX HTML 的文本
 */
export function renderMath(text) {
  if (!text) return ''

  // 收集所有数学表达式，用占位符替换，避免被其他处理干扰
  const mathBlocks = []
  let result = text

  // 1. 块级 $$...$$ 和 \[...\]
  result = result.replace(/\$\$([\s\S]*?)\$\$/g, (_, formula) => {
    const idx = mathBlocks.length
    try {
      mathBlocks.push(katex.renderToString(formula.trim(), { displayMode: true, throwOnError: false }))
    } catch { mathBlocks.push(formula) }
    return `\x00MATH${idx}\x00`
  })
  result = result.replace(/\\\[([\s\S]*?)\\\]/g, (_, formula) => {
    const idx = mathBlocks.length
    try {
      mathBlocks.push(katex.renderToString(formula.trim(), { displayMode: true, throwOnError: false }))
    } catch { mathBlocks.push(formula) }
    return `\x00MATH${idx}\x00`
  })

  // 2. 行内 $...$ 和 \(...\)
  result = result.replace(/\$([^$]+?)\$/g, (_, formula) => {
    const idx = mathBlocks.length
    try {
      mathBlocks.push(katex.renderToString(formula.trim(), { displayMode: false, throwOnError: false }))
    } catch { mathBlocks.push(formula) }
    return `\x00MATH${idx}\x00`
  })
  result = result.replace(/\\\((.+?)\\\)/g, (_, formula) => {
    const idx = mathBlocks.length
    try {
      mathBlocks.push(katex.renderToString(formula.trim(), { displayMode: false, throwOnError: false }))
    } catch { mathBlocks.push(formula) }
    return `\x00MATH${idx}\x00`
  })

  // 3. 还原占位符为 KaTeX HTML
  result = result.replace(/\x00MATH(\d+)\x00/g, (_, idx) => mathBlocks[parseInt(idx)] || '')

  return result
}

/**
 * 先提取数学公式 → 对非数学部分执行 HTML 转义+搜索高亮 → 再还原数学
 * 用于 Cards.vue 的 highlight() 场景（搜索高亮 + LaTeX 共存）
 * @param {string} text - 原始文本
 * @param {function} escapeAndHighlight - 对纯文本做转义+高亮的函数 (text) => html
 * @returns {string}
 */
export function renderMathWithHighlight(text, escapeAndHighlight) {
  if (!text) return ''

  // 拆分：交替的 [普通文本, 数学公式, 普通文本, ...]
  const parts = []
  const regex = /(\$\$[\s\S]*?\$\$|\\\[[\s\S]*?\\\]|\$[^$]+?\$|\\\(.+?\\\))/g
  let lastIndex = 0
  let match

  while ((match = regex.exec(text)) !== null) {
    if (match.index > lastIndex) {
      parts.push({ type: 'text', content: text.slice(lastIndex, match.index) })
    }
    parts.push({ type: 'math', content: match[0] })
    lastIndex = regex.lastIndex
  }
  if (lastIndex < text.length) {
    parts.push({ type: 'text', content: text.slice(lastIndex) })
  }

  // 分别处理
  return parts.map(part => {
    if (part.type === 'text') {
      return escapeAndHighlight(part.content)
    }
    return renderMath(part.content)
  }).join('')
}
