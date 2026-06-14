import { reactive } from 'vue'

const toasts = reactive([])
let nextId = 0

export function useToast() {
  function addToast(message, type = 'info', duration = 3000) {
    const id = nextId++
    toasts.push({ id, message, type, visible: true })
    if (duration > 0) {
      setTimeout(() => removeToast(id), duration)
    }
    return id
  }

  function removeToast(id) {
    const idx = toasts.findIndex(t => t.id === id)
    if (idx !== -1) {
      toasts[idx].visible = false
      setTimeout(() => {
        const i = toasts.findIndex(t => t.id === id)
        if (i !== -1) toasts.splice(i, 1)
      }, 300)
    }
  }

  return {
    toasts,
    success: (msg, dur) => addToast(msg, 'success', dur),
    error: (msg, dur) => addToast(msg, 'error', dur),
    warning: (msg, dur) => addToast(msg, 'warning', dur),
    info: (msg, dur) => addToast(msg, 'info', dur),
    remove: removeToast,
  }
}

/**
 * confirm 替代：返回 Promise<boolean>，用 Toast 风格的模态弹窗
 * 注意：state 是模块级单例，确保 Toast.vue 和调用方共享同一份状态
 */
const confirmState = reactive({ show: false, message: '', resolve: null })

export function useConfirm() {
  function confirm(message) {
    return new Promise((resolve) => {
      confirmState.show = true
      confirmState.message = message
      confirmState.resolve = resolve
    })
  }

  function onConfirm() {
    confirmState.show = false
    confirmState.resolve?.(true)
  }

  function onCancel() {
    confirmState.show = false
    confirmState.resolve?.(false)
  }

  return { confirmState, confirm, onConfirm, onCancel }
}
