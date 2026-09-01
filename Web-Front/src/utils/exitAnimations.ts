/**
 * 「优雅退场」动画工具（美化工程 · 优雅退场）
 *
 * 在真实的数据移除（删除/归档/关闭）发生前，先为对应 DOM 元素播放退场动画，
 * 动画结束后再继续执行原有逻辑，让用户明确感知操作已完成。
 *
 * 遵循 prefers-reduced-motion：用户偏好减少动效时直接跳过动画。
 */

function prefersReducedMotion(): boolean {
  return typeof window !== 'undefined' && window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

/**
 * 为元素播放指定退场动画类，动画结束后 resolve。
 * @param el 目标元素（可为 null，安全跳过）
 * @param cls 动画类名（如 delete-out）
 * @param duration 动画时长（ms），用于兜底超时
 */
export function playExit(el: Element | null | undefined, cls: string, duration: number): Promise<void> {
  return new Promise((resolve) => {
    if (!el || prefersReducedMotion()) {
      resolve();
      return;
    }
    // 防止重复触发：先移除再强制回流后添加，确保动画可重放
    el.classList.remove(cls);
    void (el as HTMLElement).offsetWidth;
    el.classList.add(cls);
    let done = false;
    const finish = () => {
      if (done) return;
      done = true;
      el.classList.remove(cls);
      resolve();
    };
    el.addEventListener('animationend', finish, { once: true });
    setTimeout(finish, duration + 100);
  });
}

/** 任务归档：卡片折叠 + 淡出（400ms，从底部向上折叠 + 纸张翻动视觉代替） */
export function playArchiveFold(el: Element | null | undefined): Promise<void> {
  return playExit(el, 'archive-fold-out', 400);
}

/** 删除操作：红色闪烁 → 缩小 → 淡出三步退场（600ms） */
export function playDeleteOut(el: Element | null | undefined): Promise<void> {
  return playExit(el, 'delete-out', 600);
}

/** 会话关闭：从右下角收缩为一个小点（300ms） */
export function playChatCollapse(el: Element | null | undefined): Promise<void> {
  return playExit(el, 'chat-collapse', 300);
}
