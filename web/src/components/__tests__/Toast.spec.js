import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { nextTick } from 'vue'
import Toast from '../Toast.vue'
import { useToast, useConfirm } from '../../composables/useToast'

describe('Toast.vue', () => {
  let wrapper = null

  beforeEach(() => {
    // 清理所有残留的 toast
    const { toasts } = useToast()
    toasts.splice(0, toasts.length)
    // 重置 confirm 状态
    const { confirmState } = useConfirm()
    confirmState.show = false
    confirmState.message = ''
    confirmState.resolve = null
  })

  afterEach(() => {
    // 正确卸载组件（清理 Teleport 内容）
    if (wrapper) {
      wrapper.unmount()
      wrapper = null
    }
    vi.restoreAllMocks()
  })

  // ─── Toast 渲染 ───────────────────────────────

  describe('Toast 通知渲染', () => {
    it('初始状态不显示任何 toast', () => {
      wrapper = mount(Toast)
      const items = wrapper.findAll('.pointer-events-auto')
      expect(items).toHaveLength(0)
    })

    it('success 类型显示 ✓ 图标和绿色样式', async () => {
      const { success } = useToast()
      wrapper = mount(Toast)

      success('操作成功', 0) // duration=0 不自动消失
      await nextTick()

      const items = wrapper.findAll('.pointer-events-auto')
      expect(items).toHaveLength(1)
      expect(items[0].text()).toContain('✓')
      expect(items[0].text()).toContain('操作成功')
      // 验证绿色样式类
      expect(items[0].classes().join(' ')).toContain('bg-green-50')
    })

    it('error 类型显示 ✕ 图标和红色样式', async () => {
      const { error } = useToast()
      wrapper = mount(Toast)

      error('出错了', 0)
      await nextTick()

      const items = wrapper.findAll('.pointer-events-auto')
      expect(items).toHaveLength(1)
      expect(items[0].text()).toContain('✕')
      expect(items[0].text()).toContain('出错了')
      expect(items[0].classes().join(' ')).toContain('bg-red-50')
    })

    it('warning 类型显示 ⚠ 图标和琥珀色样式', async () => {
      const { warning } = useToast()
      wrapper = mount(Toast)

      warning('请注意', 0)
      await nextTick()

      const items = wrapper.findAll('.pointer-events-auto')
      expect(items).toHaveLength(1)
      expect(items[0].text()).toContain('⚠')
      expect(items[0].text()).toContain('请注意')
      expect(items[0].classes().join(' ')).toContain('bg-amber-50')
    })

    it('info 类型显示 ℹ 图标和蓝色样式', async () => {
      const { info } = useToast()
      wrapper = mount(Toast)

      info('提示信息', 0)
      await nextTick()

      const items = wrapper.findAll('.pointer-events-auto')
      expect(items).toHaveLength(1)
      expect(items[0].text()).toContain('ℹ')
      expect(items[0].text()).toContain('提示信息')
      expect(items[0].classes().join(' ')).toContain('bg-blue-50')
    })

    it('同时显示多个 toast', async () => {
      const { success, error, info } = useToast()
      wrapper = mount(Toast)

      success('第一个', 0)
      error('第二个', 0)
      info('第三个', 0)
      await nextTick()

      const items = wrapper.findAll('.pointer-events-auto')
      expect(items).toHaveLength(3)
    })

    it('点击关闭按钮移除 toast', async () => {
      vi.useFakeTimers()
      const { success } = useToast()
      wrapper = mount(Toast)

      success('要移除的', 0)
      await nextTick()

      expect(wrapper.findAll('.pointer-events-auto')).toHaveLength(1)

      // 点击关闭按钮
      const closeBtn = wrapper.find('.pointer-events-auto button')
      await closeBtn.trigger('click')
      await nextTick()

      // 触发 300ms 延迟删除
      vi.advanceTimersByTime(350)
      await nextTick()

      expect(wrapper.findAll('.pointer-events-auto')).toHaveLength(0)
      vi.useRealTimers()
    })

    it('自动超时移除 toast', async () => {
      vi.useFakeTimers()
      const { info } = useToast()
      wrapper = mount(Toast)

      info('自动消失', 1000)
      await nextTick()
      expect(wrapper.findAll('.pointer-events-auto')).toHaveLength(1)

      // 前进到超时
      vi.advanceTimersByTime(1000)
      await nextTick()

      // 再前进 300ms 移除动画
      vi.advanceTimersByTime(350)
      await nextTick()

      expect(wrapper.findAll('.pointer-events-auto')).toHaveLength(0)
      vi.useRealTimers()
    })
  })

  // ─── Confirm 弹窗 ───────────────────────────────

  describe('Confirm 确认弹窗', () => {
    // Confirm 使用 <Teleport to="body">，内容在 document.body 中
    // 用 document.body 查询 Teleport 渲染的内容

    it('初始状态不显示弹窗', () => {
      wrapper = mount(Toast)
      const overlay = document.body.querySelector('.fixed.inset-0')
      expect(overlay).toBeNull()
    })

    it('调用 confirm 后显示弹窗', async () => {
      const { confirm } = useConfirm()
      wrapper = mount(Toast)

      const promise = confirm('确认删除？')
      await flushPromises()
      await nextTick()

      const overlay = document.body.querySelector('.fixed.inset-0')
      expect(overlay).not.toBeNull()
      expect(overlay.textContent).toContain('确认删除？')

      // 清理：点击取消
      const cancelBtn = overlay.querySelectorAll('button')[0]
      cancelBtn.click()
      await flushPromises()
    })

    it('点击确定按钮返回 true', async () => {
      const { confirm } = useConfirm()
      wrapper = mount(Toast)

      const promise = confirm('确认操作？')
      await flushPromises()
      await nextTick()

      const overlay = document.body.querySelector('.fixed.inset-0')
      const buttons = overlay.querySelectorAll('button')
      expect(buttons).toHaveLength(2)
      expect(buttons[1].textContent).toBe('确定')
      buttons[1].click()

      const result = await promise
      expect(result).toBe(true)
    })

    it('点击取消按钮返回 false', async () => {
      const { confirm } = useConfirm()
      wrapper = mount(Toast)

      const promise = confirm('确认操作？')
      await flushPromises()
      await nextTick()

      const overlay = document.body.querySelector('.fixed.inset-0')
      const buttons = overlay.querySelectorAll('button')
      expect(buttons[0].textContent).toBe('取消')
      buttons[0].click()

      const result = await promise
      expect(result).toBe(false)
    })

    it('Esc 键关闭确认弹窗', async () => {
      const { confirm } = useConfirm()
      wrapper = mount(Toast)

      const promise = confirm('按 Esc 关闭')
      await flushPromises()
      await nextTick()

      // 模拟 Esc 键
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
      await flushPromises()
      await nextTick()

      const result = await promise
      expect(result).toBe(false)
    })
  })

  // ─── toastClasses 函数 ────────────────────────────

  describe('toastClasses 颜色映射', () => {
    it('暗色模式 dark: 变体包含在 CSS 类中', async () => {
      const { success, error, warning, info } = useToast()
      wrapper = mount(Toast)

      success('s', 0)
      error('e', 0)
      warning('w', 0)
      info('i', 0)
      await nextTick()

      const items = wrapper.findAll('.pointer-events-auto')
      expect(items).toHaveLength(4)

      // 每个类型都有 dark: 变体
      expect(items[0].classes().join(' ')).toContain('dark:bg-green-900/30')
      expect(items[1].classes().join(' ')).toContain('dark:bg-red-900/30')
      expect(items[2].classes().join(' ')).toContain('dark:bg-amber-900/30')
      expect(items[3].classes().join(' ')).toContain('dark:bg-blue-900/30')
    })
  })
})
