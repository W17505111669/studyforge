import { ref, reactive } from 'vue'
import { explainConcept } from '../api/client'

// 模块级单例状态（多组件共享）
const selectedText = ref('')
const buttonPos = reactive({ x: 0, y: 0, visible: false })
const drawer = reactive({
  visible: false,
  loading: false,
  result: null,
  error: null
})

let mouseUpHandler = null

/**
 * useExplain — AI 概念解释器 composable
 * 提供双击选中文字后浮出"AI 解释"按钮+侧边抽屉展示结果
 */
export function useExplain() {
  // 处理 mouseup 事件：检测文字选中
  function handleMouseUp(e) {
    // 忽略抽屉内的选中
    if (e.target.closest?.('.explain-drawer')) return
    // 忽略按钮上的点击
    if (e.target.closest?.('.explain-float-btn')) return

    const sel = window.getSelection()
    const text = sel?.toString().trim()

    if (text && text.length >= 1 && text.length <= 200) {
      selectedText.value = text
      // 计算浮动按钮位置（选中区域右上角）
      const range = sel.getRangeAt(0)
      const rect = range.getBoundingClientRect()
      buttonPos.x = Math.min(rect.right + 8, window.innerWidth - 120)
      buttonPos.y = Math.max(rect.top - 8, 8)
      buttonPos.visible = true
    } else {
      hideButton()
    }
  }

  function hideButton() {
    buttonPos.visible = false
    selectedText.value = ''
  }

  // 初始化全局 mouseup 监听
  function initExplainListener() {
    mouseUpHandler = handleMouseUp
    document.addEventListener('mouseup', mouseUpHandler)
    document.addEventListener('touchend', mouseUpHandler)
  }

  function cleanupExplainListener() {
    if (mouseUpHandler) {
      document.removeEventListener('mouseup', mouseUpHandler)
      document.removeEventListener('touchend', mouseUpHandler)
      mouseUpHandler = null
    }
  }

  // 请求 AI 解释
  async function requestExplain(concept, context) {
    drawer.visible = true
    drawer.loading = true
    drawer.error = null
    drawer.result = null
    hideButton()

    try {
      const res = await explainConcept({ concept, context: context || '' })
      drawer.result = res.data
    } catch (err) {
      drawer.error = err.response?.data?.error || '解释生成失败，请稍后重试'
    } finally {
      drawer.loading = false
    }
  }

  // 用选中文字触发解释
  function explainSelection() {
    if (selectedText.value) {
      requestExplain(selectedText.value)
    }
  }

  function closeDrawer() {
    drawer.visible = false
    drawer.result = null
    drawer.error = null
  }

  return {
    selectedText,
    buttonPos,
    drawer,
    hideButton,
    initExplainListener,
    cleanupExplainListener,
    requestExplain,
    explainSelection,
    closeDrawer
  }
}
