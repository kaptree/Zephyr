package utils

import (
	"strings"
	"sync"
	"unicode"

	"github.com/mozillazg/go-pinyin"
)

// IsPinyinKeyword 判断关键词是否为纯 ASCII（拼音/英文搜索场景）
func IsPinyinKeyword(kw string) bool {
	if kw == "" {
		return false
	}
	for _, r := range kw {
		if r > 127 {
			return false
		}
	}
	return true
}

var (
	pinyinOnce  sync.Once
	pinyinArgs  pinyin.Args
	pinyinMu    sync.RWMutex
	pinyinCache = map[string][2]string{} // 原文 -> [全拼, 拼音首字母]
)

// pinyinVariants 返回文本的全拼与拼音首字母（小写，带缓存）。
// 连续汉字按词组转换（go-pinyin 词组库处理常见多音字），非汉字字符原样保留。
func pinyinVariants(text string) (string, string) {
	pinyinMu.RLock()
	v, ok := pinyinCache[text]
	pinyinMu.RUnlock()
	if ok {
		return v[0], v[1]
	}

	pinyinOnce.Do(func() {
		pinyinArgs = pinyin.NewArgs()
		pinyinArgs.Style = pinyin.Normal // 无声调
		pinyinArgs.Heteronym = false
	})

	var full, initials strings.Builder
	var hanSeg strings.Builder
	flushHan := func() {
		if hanSeg.Len() == 0 {
			return
		}
		for _, p := range pinyin.LazyPinyin(hanSeg.String(), pinyinArgs) {
			lp := strings.ToLower(p)
			full.WriteString(lp)
			if lp != "" {
				initials.WriteString(lp[:1])
			}
		}
		hanSeg.Reset()
	}
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			hanSeg.WriteRune(r)
			continue
		}
		flushHan()
		lr := unicode.ToLower(r)
		full.WriteRune(lr)
		initials.WriteRune(lr)
	}
	flushHan()

	res := [2]string{full.String(), initials.String()}
	pinyinMu.Lock()
	pinyinCache[text] = res
	pinyinMu.Unlock()
	return res[0], res[1]
}

// MatchKeyword 判断任一候选文本是否命中关键词：
// 原文（忽略大小写）、拼音全拼、拼音首字母 任一包含即命中
func MatchKeyword(keyword string, texts ...string) bool {
	kw := strings.ToLower(strings.TrimSpace(keyword))
	if kw == "" {
		return true
	}
	for _, t := range texts {
		if t == "" {
			continue
		}
		if strings.Contains(strings.ToLower(t), kw) {
			return true
		}
		full, initials := pinyinVariants(t)
		if strings.Contains(full, kw) || strings.Contains(initials, kw) {
			return true
		}
	}
	return false
}
