/**
 * 系统提示音：使用 Web Audio API 合成双音提示音，无需外部资源
 */
let audioCtx: AudioContext | null = null;

function getCtx(): AudioContext | null {
  if (typeof window === 'undefined') return null;
  try {
    if (!audioCtx) {
      const Ctor =
        window.AudioContext ||
        (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
      if (!Ctor) return null;
      audioCtx = new Ctor();
    }
    if (audioCtx.state === 'suspended') {
      void audioCtx.resume();
    }
    return audioCtx;
  } catch {
    return null;
  }
}

function tone(ctx: AudioContext, freq: number, start: number, duration: number, volume = 0.16) {
  const osc = ctx.createOscillator();
  const gain = ctx.createGain();
  osc.type = 'sine';
  osc.frequency.value = freq;
  gain.gain.setValueAtTime(0, ctx.currentTime + start);
  gain.gain.linearRampToValueAtTime(volume, ctx.currentTime + start + 0.02);
  gain.gain.exponentialRampToValueAtTime(0.0001, ctx.currentTime + start + duration);
  osc.connect(gain);
  gain.connect(ctx.destination);
  osc.start(ctx.currentTime + start);
  osc.stop(ctx.currentTime + start + duration + 0.05);
}

/** 新通知提示音（清脆双音） */
export function playNotificationSound() {
  const ctx = getCtx();
  if (!ctx) return;
  tone(ctx, 880, 0, 0.18);
  tone(ctx, 1174.66, 0.12, 0.22);
}

/** 新聊天消息提示音（柔和单音） */
export function playChatSound() {
  const ctx = getCtx();
  if (!ctx) return;
  tone(ctx, 660, 0, 0.14, 0.12);
}
