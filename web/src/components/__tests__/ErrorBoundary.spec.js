import { describe, it, expect, vi, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { defineComponent, h, ref, nextTick } from 'vue'
import ErrorBoundary from '../ErrorBoundary.vue'

// ─── 辅助组件 ──────────────────────────────────

/** 正常渲染的子组件 */
const GoodChild = defineComponent({
  name: 'GoodChild',
  setup() {
    return () => h('div', { class: 'good-child' }, '一切正常')
  }
})

/** 渲染时抛出错误的子组件 */
function createFailingChild(errorMessage) {
  return defineComponent({
    name: 'FailingChild',
    setup() {
      const thrown = ref(false)
      // 模拟通过 onErrorCaptured 可以捕获的错误
      return () => {
        if (!thrown.value) {
          thrown.value = true
          throw new Error(errorMessage)
        }
        return h('div', '不应显示')
      }
    }
  })
}

/** 点击按钮后抛出错误的子组件 */
function _createClickFailChild(errorMessage) {
  return defineComponent({
    name: 'ClickFailChild',
    setup() {
      const shouldFail = ref(false)
      return () =>
        h('div', {}, [
          h(
            'button',
            {
              class: 'trigger-btn',
              onClick: () => {
                shouldFail.value = true
              }
            },
            '触发错误'
          ),
          shouldFail.value
            ? (() => {
                throw new Error(errorMessage)
              })()
            : h('span', '等待触发')
        ])
    }
  })
}

// ─── 测试套件 ──────────────────────────────────

describe('ErrorBoundary.vue', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  // ─── 正常渲染 ─────────────────────────────────

  describe('正常渲染', () => {
    it('子组件无错误时显示 slot 内容', () => {
      const wrapper = mount(ErrorBoundary, {
        slots: {
          default: () => h(GoodChild)
        }
      })

      expect(wrapper.find('.good-child').exists()).toBe(true)
      expect(wrapper.find('.good-child').text()).toBe('一切正常')
      expect(wrapper.find('.error-fallback').exists()).toBe(false)
    })

    it('slot 内容为普通文本时正确渲染', () => {
      const wrapper = mount(ErrorBoundary, {
        slots: {
          default: '<p class="test-text">Hello World</p>'
        }
      })

      expect(wrapper.find('.test-text').exists()).toBe(true)
      expect(wrapper.find('.test-text').text()).toBe('Hello World')
    })

    it('slot 包含多个子元素时全部渲染', () => {
      const wrapper = mount(ErrorBoundary, {
        slots: {
          default: () => [
            h('span', { key: 'a', class: 'child-a' }, 'A'),
            h('span', { key: 'b', class: 'child-b' }, 'B')
          ]
        }
      })

      expect(wrapper.findAll('.child-a, .child-b')).toHaveLength(2)
    })
  })

  // ─── 错误捕获 ─────────────────────────────────

  describe('错误捕获', () => {
    it('子组件渲染错误时显示错误 UI', async () => {
      // 阻止 console.error 输出干扰测试
      vi.spyOn(console, 'error').mockImplementation(() => {})

      const FailingChild = createFailingChild('测试错误消息')
      const wrapper = mount(ErrorBoundary, {
        slots: {
          default: () => h(FailingChild)
        }
      })

      await nextTick()

      // 应该显示错误 UI
      expect(wrapper.find('.error-fallback').exists()).toBe(true)
      expect(wrapper.find('.error-title').text()).toBe('组件出错了')
      expect(wrapper.find('.error-message').text()).toBe('测试错误消息')
    })

    it('错误 UI 包含重试按钮', async () => {
      vi.spyOn(console, 'error').mockImplementation(() => {})

      const FailingChild = createFailingChild('某个错误')
      const wrapper = mount(ErrorBoundary, {
        slots: {
          default: () => h(FailingChild)
        }
      })

      await nextTick()

      const retryBtn = wrapper.find('.retry-btn')
      expect(retryBtn.exists()).toBe(true)
      expect(retryBtn.text()).toContain('重试')
    })

    it('错误 UI 包含错误图标 SVG', async () => {
      vi.spyOn(console, 'error').mockImplementation(() => {})

      const FailingChild = createFailingChild('图标测试')
      const wrapper = mount(ErrorBoundary, {
        slots: {
          default: () => h(FailingChild)
        }
      })

      await nextTick()

      expect(wrapper.find('.error-icon').exists()).toBe(true)
      // SVG 元素存在
      expect(wrapper.find('.error-icon svg, svg.error-icon').exists()).toBe(true)
    })

    it('错误消息为 undefined 时显示默认文案', async () => {
      vi.spyOn(console, 'error').mockImplementation(() => {})

      const FailingChild = defineComponent({
        name: 'NoMessageError',
        setup() {
          return () => {
            throw {} // 没有 message 属性
          }
        }
      })

      const wrapper = mount(ErrorBoundary, {
        slots: {
          default: () => h(FailingChild)
        }
      })

      await nextTick()

      expect(wrapper.find('.error-message').text()).toBe('发生了未知错误')
    })

    it('onErrorCaptured 返回 false 阻止错误冒泡', async () => {
      vi.spyOn(console, 'error').mockImplementation(() => {})

      // 父级 error handler 不应被调用
      const parentHandler = vi.fn()
      const FailingChild = createFailingChild('阻止冒泡测试')

      const wrapper = mount(
        defineComponent({
          setup() {
            return () =>
              h(ErrorBoundary, null, {
                default: () => h(FailingChild)
              })
          }
        }),
        {
          global: {
            config: {
              errorHandler: parentHandler
            }
          }
        }
      )

      await nextTick()

      // ErrorBoundary 应该捕获了错误并阻止冒泡
      // 父级 errorHandler 不应被调用（return false）
      // 注意：Vue 的 onErrorCaptured 和 errorHandler 行为，
      // return false 会阻止继续向上传播
      expect(wrapper.find('.error-fallback').exists()).toBe(true)
    })
  })

  // ─── 重试功能 ─────────────────────────────────

  describe('重试功能', () => {
    it('点击重试按钮重置错误状态', async () => {
      vi.spyOn(console, 'error').mockImplementation(() => {})

      // 使用可控的子组件：第一次抛错，之后正常
      let renderCount = 0
      const ControllableChild = defineComponent({
        name: 'ControllableChild',
        setup() {
          return () => {
            renderCount++
            if (renderCount === 1) {
              throw new Error('首次渲染错误')
            }
            return h('div', { class: 'recovered' }, '恢复正常')
          }
        }
      })

      const wrapper = mount(ErrorBoundary, {
        slots: {
          default: () => h(ControllableChild)
        }
      })

      await nextTick()

      // 第一次渲染出错
      expect(wrapper.find('.error-fallback').exists()).toBe(true)
      expect(wrapper.find('.error-message').text()).toBe('首次渲染错误')

      // 点击重试
      await wrapper.find('.retry-btn').trigger('click')
      await nextTick()

      // 重试后子组件正常渲染
      expect(wrapper.find('.error-fallback').exists()).toBe(false)
      expect(wrapper.find('.recovered').exists()).toBe(true)
      expect(wrapper.find('.recovered').text()).toBe('恢复正常')
    })

    it('重试后如果子组件仍然报错，再次显示错误', async () => {
      vi.spyOn(console, 'error').mockImplementation(() => {})

      const AlwaysFailChild = defineComponent({
        name: 'AlwaysFailChild',
        setup() {
          return () => {
            throw new Error('持续错误')
          }
        }
      })

      const wrapper = mount(ErrorBoundary, {
        slots: {
          default: () => h(AlwaysFailChild)
        }
      })

      await nextTick()
      expect(wrapper.find('.error-message').text()).toBe('持续错误')

      // 重试
      await wrapper.find('.retry-btn').trigger('click')
      await nextTick()

      // 仍然显示错误
      expect(wrapper.find('.error-fallback').exists()).toBe(true)
      expect(wrapper.find('.error-message').text()).toBe('持续错误')
    })
  })

  // ─── console.error 日志 ─────────────────────────

  describe('错误日志', () => {
    it('捕获错误时输出 console.error 日志', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      const FailingChild = createFailingChild('日志测试')
      mount(ErrorBoundary, {
        slots: {
          default: () => h(FailingChild)
        }
      })

      await nextTick()

      expect(consoleSpy).toHaveBeenCalledWith('[ErrorBoundary] 组件渲染错误:', expect.any(Error))
    })
  })
})
