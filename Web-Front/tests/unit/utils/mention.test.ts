import { describe, it, expect } from 'vitest';
import {
  memberDisplayName,
  buildMentionInsert,
  findActiveMentionToken,
  splitMentionSegments,
  extractMentionedNames,
  renderMessageContent,
  filterMentionMembers,
} from '@/utils/mention';
import type { GroupMemberItem } from '@/types';

function member(overrides: Partial<GroupMemberItem> & { user_id: string }): GroupMemberItem {
  return {
    id: overrides.user_id,
    group_id: 'g1',
    role: 'member',
    joined_at: '2026-09-07T00:00:00Z',
    ...overrides,
  } as GroupMemberItem;
}

describe('memberDisplayName', () => {
  it('优先级：昵称 > 用户名 > 用户 ID 前 6 位', () => {
    expect(memberDisplayName({ user_id: 'uid-123456', user: { id: 'u', username: 'admin', name: '张三' } } as GroupMemberItem)).toBe('张三');
    expect(memberDisplayName({ user_id: 'uid-123456', user: { id: 'u', username: 'admin' } } as GroupMemberItem)).toBe('admin');
    expect(memberDisplayName({ user_id: 'uid-123456' } as GroupMemberItem)).toBe('uid-12');
  });
});

describe('buildMentionInsert', () => {
  it('插入文本为 @名字 + 空格（空格用于结束提及）', () => {
    expect(buildMentionInsert('张三')).toBe('@张三 ');
  });
});

describe('findActiveMentionToken', () => {
  it('表驱动：合法/非法活动 @ 记号', () => {
    const cases: Array<{
      name: string;
      content: string;
      caret: number;
      want: { at: number; keyword: string } | null;
    }> = [
      { name: '光标在开头', content: '@', caret: 0, want: null },
      { name: '无 @', content: 'hello', caret: 5, want: null },
      { name: '开头 @', content: '@张', caret: 2, want: { at: 0, keyword: '张' } },
      { name: '空白后 @', content: 'hi @张', caret: 5, want: { at: 3, keyword: '张' } },
      { name: '邮箱场景不触发', content: 'a@b.com', caret: 7, want: null },
      { name: '过滤词含空白不触发', content: '@张 三', caret: 5, want: null },
      { name: '过滤词含另一个 @ 不触发', content: '@张@李', caret: 4, want: null },
      { name: '空格结束提及后不触发', content: '@张三 ', caret: 4, want: null },
      { name: '过滤词超长不触发', content: `@${'x'.repeat(21)}`, caret: 22, want: null },
    ];
    for (const c of cases) {
      expect(findActiveMentionToken(c.content, c.caret), c.name).toEqual(c.want);
    }
  });
});

describe('splitMentionSegments', () => {
  it('空文本返回空数组', () => {
    expect(splitMentionSegments('', ['张三'])).toEqual([]);
  });

  it('无候选名时整段不提及', () => {
    expect(splitMentionSegments('hello @张三', [])).toEqual([
      { text: 'hello @张三', mentioned: false },
    ]);
  });

  it('表驱动：有效提及分段', () => {
    const names = ['张三', '李四'];
    const cases: Array<{ name: string; content: string; want: Array<{ text: string; mentioned: boolean }> }> = [
      {
        name: '开头单个提及',
        content: '@张三',
        want: [{ text: '@张三', mentioned: true }],
      },
      {
        name: '多个提及（空格分隔）',
        content: '大家好 @张三 和 @李四',
        want: [
          { text: '大家好 ', mentioned: false },
          { text: '@张三', mentioned: true },
          { text: ' 和 ', mentioned: false },
          { text: '@李四', mentioned: true },
        ],
      },
      {
        name: '非候选名不视为提及',
        content: '@王五 你好',
        want: [{ text: '@王五 你好', mentioned: false }],
      },
      {
        name: '@ 前无空白（如邮箱）不视为提及',
        content: '看@张三',
        want: [{ text: '看@张三', mentioned: false }],
      },
      {
        name: '紧邻两个 @ 仅首个有效（无空格结尾不连续）',
        content: '@张三@李四',
        want: [{ text: '@张三', mentioned: true }, { text: '@李四', mentioned: false }],
      },
      {
        name: '换行后提及有效',
        content: '你好\n@张三',
        want: [
          { text: '你好\n', mentioned: false },
          { text: '@张三', mentioned: true },
        ],
      },
    ];
    for (const c of cases) {
      expect(splitMentionSegments(c.content, names), c.name).toEqual(c.want);
    }
  });
});

describe('extractMentionedNames', () => {
  it('仅提取候选名，去重保序', () => {
    expect(extractMentionedNames('hi @张三 @王五 @张三 @李四', ['张三', '李四'])).toEqual([
      '张三',
      '李四',
    ]);
  });

  it('无有效提及返回空数组', () => {
    expect(extractMentionedNames('邮箱 a@b.com', ['张三'])).toEqual([]);
  });
});

describe('renderMessageContent', () => {
  it('空文本返回空字符串', () => {
    expect(renderMessageContent('', ['张三'])).toBe('');
  });

  it('含 HTML 标签视为富文本原样返回', () => {
    const html = '<b>hi</b>';
    expect(renderMessageContent(html, ['张三'])).toBe(html);
  });

  it('@名字 包 mention-tag 标记，其余文本正常转义', () => {
    expect(renderMessageContent('hi @张三', ['张三'])).toBe(
      'hi <span class="mention-tag">@张三</span>'
    );
  });

  it('换行转 <br>，提及段内特殊字符转义', () => {
    expect(renderMessageContent('你好\n@张&三', ['张&三'])).toBe(
      '你好<br><span class="mention-tag">@张&amp;三</span>'
    );
  });

  it('非候选名不包标记但仍转义', () => {
    expect(renderMessageContent('@王五 <提示>', ['张三'])).toBe('@王五 &lt;提示&gt;');
  });
});

describe('filterMentionMembers', () => {
  const members: GroupMemberItem[] = [
    member({ user_id: 'u1', user: { id: 'u1', username: 'zhangsan', name: '张三' } }),
    member({
      user_id: 'u2',
      user: {
        id: 'u2',
        username: 'lisi',
        name: '李四',
        department: { id: 'd1', name: '销售部' },
      },
    }),
    member({ user_id: 'u3' }),
  ];

  it('空关键字返回全部成员', () => {
    expect(filterMentionMembers(members, '')).toHaveLength(3);
    expect(filterMentionMembers(members, '   ')).toHaveLength(3);
  });

  it('按昵称原文匹配', () => {
    expect(filterMentionMembers(members, '张三').map((m) => m.user_id)).toEqual(['u1']);
  });

  it('按用户名匹配', () => {
    expect(filterMentionMembers(members, 'lisi').map((m) => m.user_id)).toEqual(['u2']);
  });

  it('按部门名匹配', () => {
    expect(filterMentionMembers(members, '销售部').map((m) => m.user_id)).toEqual(['u2']);
  });

  it('支持拼音首字母匹配（zs → 张三）', () => {
    expect(filterMentionMembers(members, 'zs').map((m) => m.user_id)).toEqual(['u1']);
  });

  it('无 user 信息的成员按 ID 兜底参与匹配', () => {
    expect(filterMentionMembers(members, 'u3').map((m) => m.user_id)).toEqual(['u3']);
  });

  it('无命中返回空数组', () => {
    expect(filterMentionMembers(members, 'wangwu')).toEqual([]);
  });
});
