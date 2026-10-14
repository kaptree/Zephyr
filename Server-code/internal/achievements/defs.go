package achievements

import "fmt"

// 成就分类
const (
	CategoryJourney     = "journey"     // 任务征程：创建/完成/归档里程碑
	CategoryEfficiency  = "efficiency"  // 高效达人：专注/截止/模板
	CategoryColor       = "color"       // 色彩玩家：便签四色收集
	CategorySocial      = "social"      // 社交蝴蝶：私聊/群聊/表情包
	CategoryCommand     = "command"     // 协作领袖：指派/抄送/反馈/盯办
	CategoryHunter      = "hunter"      // Bug 猎手：问题反馈
	CategoryExplorer    = "explorer"    // AI 探索：报告生成
	CategoryPersonality = "personality" // 个性陪伴：装扮/坚持/彩蛋
)

// Category 成就分类元信息
type Category struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Icon string `json:"icon"` // 前端图标名
	Desc string `json:"desc"`
}

// Categories 分类展示顺序
var Categories = []Category{
	{CategoryJourney, "任务征程", "clipboard", "从第一条便签到任务帝国"},
	{CategoryEfficiency, "高效达人", "clock", "专注与时间的主人"},
	{CategoryColor, "色彩玩家", "star", "收集便签的每一种颜色"},
	{CategorySocial, "社交蝴蝶", "chat", "让沟通连接每一个人"},
	{CategoryCommand, "协作领袖", "users", "团队运作的幕后指挥"},
	{CategoryHunter, "Bug 猎手", "bug", "发现并消灭每一个问题"},
	{CategoryExplorer, "AI 探索", "monitor", "与 AI 一起总结工作"},
	{CategoryPersonality, "个性与陪伴", "user", "装扮自己，坚持热爱"},
}

// Def 成就定义
type Def struct {
	Code       string `json:"code"`         // 唯一编码
	Name       string `json:"name"`         // 成就名
	Desc       string `json:"desc"`         // 达成条件描述
	Category   string `json:"category"`     // 分类 key
	Icon       string `json:"icon"`         // 前端图标名
	Tier       string `json:"tier"`         // bronze/silver/gold/platinum 稀有度
	Target     int64  `json:"target"`       // 达成阈值（对 metric 计数）
	IsMystery  bool   `json:"is_mystery"`   // 彩蛋成就：解锁前不显示达成条件
}

// defs 按 code 索引的全量定义（恰好 100 个）
var defs = func() map[string]Def {
	m := make(map[string]Def, len(defList))
	for _, d := range defList {
		if _, dup := m[d.Code]; dup {
			panic(fmt.Sprintf("duplicate achievement code: %s", d.Code))
		}
		m[d.Code] = d
	}
	return m
}()

// Defs 返回全部成就定义
func Defs() map[string]Def { return defs }

// DefList 返回有序成就列表
func DefList() []Def { return defList }

// Get 按 code 取定义
func Get(code string) (Def, bool) { d, ok := defs[code]; return d, ok }

// tier 常量
const (
	TierBronze   = "bronze"
	TierSilver   = "silver"
	TierGold     = "gold"
	TierPlatinum = "platinum"
)

// defList 全部 100 个成就定义（顺序即展示顺序）
var defList = []Def{
	// ───────────── 一、任务征程（20）─────────────
	{Code: "first_step", Name: "初出茅庐", Desc: "创建你的第 1 个任务", Category: CategoryJourney, Icon: "pencil", Tier: TierBronze, Target: 1},
	{Code: "create_10", Name: "小试牛刀", Desc: "累计创建 10 个任务", Category: CategoryJourney, Icon: "pencil", Tier: TierBronze, Target: 10},
	{Code: "create_25", Name: "渐入佳境", Desc: "累计创建 25 个任务", Category: CategoryJourney, Icon: "pencil", Tier: TierBronze, Target: 25},
	{Code: "create_50", Name: "妙笔生花", Desc: "累计创建 50 个任务", Category: CategoryJourney, Icon: "pencil", Tier: TierSilver, Target: 50},
	{Code: "create_100", Name: "百步穿杨", Desc: "累计创建 100 个任务", Category: CategoryJourney, Icon: "pencil", Tier: TierSilver, Target: 100},
	{Code: "create_250", Name: "便签工厂", Desc: "累计创建 250 个任务", Category: CategoryJourney, Icon: "pencil", Tier: TierGold, Target: 250},
	{Code: "create_500", Name: "便签海洋", Desc: "累计创建 500 个任务", Category: CategoryJourney, Icon: "pencil", Tier: TierGold, Target: 500},
	{Code: "create_1000", Name: "千军万马", Desc: "累计创建 1000 个任务", Category: CategoryJourney, Icon: "crown", Tier: TierPlatinum, Target: 1000},
	{Code: "complete_1", Name: "首战告捷", Desc: "完成你的第 1 个任务", Category: CategoryJourney, Icon: "check", Tier: TierBronze, Target: 1},
	{Code: "complete_10", Name: "十全十美", Desc: "累计完成 10 个任务", Category: CategoryJourney, Icon: "check", Tier: TierBronze, Target: 10},
	{Code: "complete_25", Name: "行云流水", Desc: "累计完成 25 个任务", Category: CategoryJourney, Icon: "check", Tier: TierBronze, Target: 25},
	{Code: "complete_50", Name: "五十连击", Desc: "累计完成 50 个任务", Category: CategoryJourney, Icon: "check", Tier: TierSilver, Target: 50},
	{Code: "complete_100", Name: "百发百中", Desc: "累计完成 100 个任务", Category: CategoryJourney, Icon: "check", Tier: TierSilver, Target: 100},
	{Code: "complete_250", Name: "收割机器", Desc: "累计完成 250 个任务", Category: CategoryJourney, Icon: "check", Tier: TierGold, Target: 250},
	{Code: "complete_500", Name: "待办清道夫", Desc: "累计完成 500 个任务", Category: CategoryJourney, Icon: "check", Tier: TierGold, Target: 500},
	{Code: "complete_1000", Name: "无情执行机", Desc: "累计完成 1000 个任务", Category: CategoryJourney, Icon: "crown", Tier: TierPlatinum, Target: 1000},
	{Code: "archive_10", Name: "收纳学徒", Desc: "归档 10 个任务", Category: CategoryJourney, Icon: "archive", Tier: TierBronze, Target: 10},
	{Code: "archive_50", Name: "整理大师", Desc: "归档 50 个任务", Category: CategoryJourney, Icon: "archive", Tier: TierSilver, Target: 50},
	{Code: "archive_100", Name: "极简主义者", Desc: "归档 100 个任务", Category: CategoryJourney, Icon: "archive", Tier: TierGold, Target: 100},
	{Code: "same_day_10", Name: "今日事今日毕", Desc: "10 个任务当天创建当天完成", Category: CategoryJourney, Icon: "clock", Tier: TierSilver, Target: 10},

	// ───────────── 二、高效达人（12）─────────────
	{Code: "worktime_1h", Name: "心流初体验", Desc: "累计专注 1 小时", Category: CategoryEfficiency, Icon: "clock", Tier: TierBronze, Target: 3600},
	{Code: "worktime_10h", Name: "深度工作者", Desc: "累计专注 10 小时", Category: CategoryEfficiency, Icon: "clock", Tier: TierSilver, Target: 36000},
	{Code: "worktime_50h", Name: "时间琥珀", Desc: "累计专注 50 小时", Category: CategoryEfficiency, Icon: "clock", Tier: TierGold, Target: 180000},
	{Code: "worktime_100h", Name: "时间管理大师", Desc: "累计专注 100 小时", Category: CategoryEfficiency, Icon: "crown", Tier: TierPlatinum, Target: 360000},
	{Code: "due_10", Name: "有备而来", Desc: "为 10 个任务设置截止时间", Category: CategoryEfficiency, Icon: "clock", Tier: TierBronze, Target: 10},
	{Code: "due_50", Name: "守时狂人", Desc: "为 50 个任务设置截止时间", Category: CategoryEfficiency, Icon: "clock", Tier: TierSilver, Target: 50},
	{Code: "pinned_5", Name: "置顶强迫症", Desc: "同时置顶 5 个任务", Category: CategoryEfficiency, Icon: "star", Tier: TierBronze, Target: 5},
	{Code: "content_5000", Name: "笔耕不辍", Desc: "累计输入 5000 字任务内容", Category: CategoryEfficiency, Icon: "pencil", Tier: TierBronze, Target: 5000},
	{Code: "content_50000", Name: "著作等身", Desc: "累计输入 50000 字任务内容", Category: CategoryEfficiency, Icon: "pencil", Tier: TierGold, Target: 50000},
	{Code: "template_1", Name: "模板尝鲜", Desc: "使用模板创建 1 个任务", Category: CategoryEfficiency, Icon: "template", Tier: TierBronze, Target: 1},
	{Code: "template_10", Name: "模板熟练工", Desc: "使用模板创建 10 个任务", Category: CategoryEfficiency, Icon: "template", Tier: TierSilver, Target: 10},
	{Code: "night_5", Name: "夜猫子", Desc: "凌晨 0-5 点创建 5 个任务", Category: CategoryEfficiency, Icon: "moon", Tier: TierSilver, Target: 5},

	// ───────────── 三、色彩玩家（6）─────────────
	{Code: "yellow_10", Name: "阳光普照", Desc: "创建 10 个黄色便签", Category: CategoryColor, Icon: "sun", Tier: TierBronze, Target: 10},
	{Code: "red_10", Name: "红色警戒", Desc: "创建 10 个红色便签", Category: CategoryColor, Icon: "warning", Tier: TierBronze, Target: 10},
	{Code: "green_10", Name: "绿意盎然", Desc: "创建 10 个绿色便签", Category: CategoryColor, Icon: "check-circle", Tier: TierBronze, Target: 10},
	{Code: "blue_10", Name: "蓝色心动", Desc: "创建 10 个蓝色便签", Category: CategoryColor, Icon: "info", Tier: TierBronze, Target: 10},
	{Code: "color_rainbow", Name: "彩虹画家", Desc: "四种颜色便签各集齐 10 个", Category: CategoryColor, Icon: "star", Tier: TierGold, Target: 1},
	{Code: "color_50", Name: "调色盘大师", Desc: "累计创建 50 个彩色便签", Category: CategoryColor, Icon: "star", Tier: TierSilver, Target: 50},

	// ───────────── 四、社交蝴蝶（15）─────────────
	{Code: "chat_1", Name: "打破沉默", Desc: "发送第 1 条私聊消息", Category: CategorySocial, Icon: "chat", Tier: TierBronze, Target: 1},
	{Code: "chat_50", Name: "话痨萌芽", Desc: "发送 50 条私聊消息", Category: CategorySocial, Icon: "chat", Tier: TierBronze, Target: 50},
	{Code: "chat_200", Name: "聊天达人", Desc: "发送 200 条私聊消息", Category: CategorySocial, Icon: "chat", Tier: TierSilver, Target: 200},
	{Code: "chat_500", Name: "社交蝴蝶", Desc: "发送 500 条私聊消息", Category: CategorySocial, Icon: "chat", Tier: TierGold, Target: 500},
	{Code: "chat_1000", Name: "人间喇叭", Desc: "发送 1000 条私聊消息", Category: CategorySocial, Icon: "crown", Tier: TierPlatinum, Target: 1000},
	{Code: "group_join_1", Name: "入群萌新", Desc: "加入第 1 个群聊", Category: CategorySocial, Icon: "users", Tier: TierBronze, Target: 1},
	{Code: "group_join_5", Name: "群聊常客", Desc: "加入 5 个群聊", Category: CategorySocial, Icon: "users", Tier: TierSilver, Target: 5},
	{Code: "group_create_1", Name: "建群达人", Desc: "创建第 1 个群聊", Category: CategorySocial, Icon: "users", Tier: TierBronze, Target: 1},
	{Code: "group_create_5", Name: "社群领袖", Desc: "创建 5 个群聊", Category: CategorySocial, Icon: "users", Tier: TierGold, Target: 5},
	{Code: "group_msg_100", Name: "灌水先锋", Desc: "群聊发送 100 条消息", Category: CategorySocial, Icon: "chat", Tier: TierSilver, Target: 100},
	{Code: "group_msg_500", Name: "气氛组组长", Desc: "群聊发送 500 条消息", Category: CategorySocial, Icon: "chat", Tier: TierGold, Target: 500},
	{Code: "group_file_1", Name: "资源分享家", Desc: "上传第 1 个群文件", Category: CategorySocial, Icon: "list", Tier: TierBronze, Target: 1},
	{Code: "emoticon_1", Name: "表情包制造机", Desc: "上传第 1 张表情包", Category: CategorySocial, Icon: "star", Tier: TierBronze, Target: 1},
	{Code: "emoticon_10", Name: "斗图之王", Desc: "上传 10 张表情包", Category: CategorySocial, Icon: "star", Tier: TierSilver, Target: 10},
	{Code: "chat_image_10", Name: "赛博摄影师", Desc: "发送 10 张图片消息", Category: CategorySocial, Icon: "view", Tier: TierSilver, Target: 10},

	// ───────────── 五、协作领袖（12）─────────────
	{Code: "assign_1", Name: "团队指挥官", Desc: "首次指派任务给他人", Category: CategoryCommand, Icon: "users", Tier: TierBronze, Target: 1},
	{Code: "assign_10", Name: "任务分发机", Desc: "指派 10 个任务给他人", Category: CategoryCommand, Icon: "users", Tier: TierSilver, Target: 10},
	{Code: "assign_50", Name: "调度中心", Desc: "指派 50 个任务给他人", Category: CategoryCommand, Icon: "crown", Tier: TierGold, Target: 50},
	{Code: "cc_1", Name: "信息枢纽", Desc: "首次抄送任务给他人", Category: CategoryCommand, Icon: "list", Tier: TierBronze, Target: 1},
	{Code: "cc_10", Name: "同步狂魔", Desc: "抄送任务 10 次", Category: CategoryCommand, Icon: "list", Tier: TierSilver, Target: 10},
	{Code: "received_1", Name: "受任于危难之际", Desc: "首次被指派任务", Category: CategoryCommand, Icon: "bell", Tier: TierBronze, Target: 1},
	{Code: "received_10", Name: "团队粘合剂", Desc: "被指派 10 个任务", Category: CategoryCommand, Icon: "bell", Tier: TierSilver, Target: 10},
	{Code: "sign_1", Name: "签收初体验", Desc: "首次签收任务", Category: CategoryCommand, Icon: "check-circle", Tier: TierBronze, Target: 1},
	{Code: "sign_10", Name: "签名狂人", Desc: "签收 10 个任务", Category: CategoryCommand, Icon: "check-circle", Tier: TierSilver, Target: 10},
	{Code: "feedback_1", Name: "反馈先锋", Desc: "首次提交任务反馈", Category: CategoryCommand, Icon: "pencil", Tier: TierBronze, Target: 1},
	{Code: "feedback_10", Name: "评审专家", Desc: "提交 10 次任务反馈", Category: CategoryCommand, Icon: "pencil", Tier: TierSilver, Target: 10},
	{Code: "remind_1", Name: "夺命连环 call", Desc: "首次盯办他人", Category: CategoryCommand, Icon: "bell", Tier: TierBronze, Target: 1},

	// ───────────── 六、Bug 猎手（10）─────────────
	{Code: "issue_1", Name: "第一只螃蟹", Desc: "提交第 1 个问题", Category: CategoryHunter, Icon: "bug", Tier: TierBronze, Target: 1},
	{Code: "issue_5", Name: "问题发现家", Desc: "提交 5 个问题", Category: CategoryHunter, Icon: "bug", Tier: TierBronze, Target: 5},
	{Code: "issue_20", Name: "Bug 猎人", Desc: "提交 20 个问题", Category: CategoryHunter, Icon: "bug", Tier: TierSilver, Target: 20},
	{Code: "issue_50", Name: "粉碎机", Desc: "提交 50 个问题", Category: CategoryHunter, Icon: "bug", Tier: TierGold, Target: 50},
	{Code: "bug_10", Name: "十连击", Desc: "提交 10 个缺陷类型问题", Category: CategoryHunter, Icon: "bug", Tier: TierSilver, Target: 10},
	{Code: "feature_5", Name: "需求分析师", Desc: "提交 5 个功能建议", Category: CategoryHunter, Icon: "list", Tier: TierSilver, Target: 5},
	{Code: "issue_comment_1", Name: "围观群众", Desc: "首次评论问题", Category: CategoryHunter, Icon: "chat", Tier: TierBronze, Target: 1},
	{Code: "issue_comment_20", Name: "热心市民", Desc: "累计评论问题 20 次", Category: CategoryHunter, Icon: "chat", Tier: TierSilver, Target: 20},
	{Code: "issue_watcher", Name: "全局守望者", Desc: "订阅全部问题的通知", Category: CategoryHunter, Icon: "bell", Tier: TierGold, Target: 1},
	{Code: "issue_multi_5", Name: "五湖四海", Desc: "在 5 个不同问题下评论", Category: CategoryHunter, Icon: "chat", Tier: TierSilver, Target: 5},

	// ───────────── 七、AI 探索（8）─────────────
	{Code: "report_1", Name: "报告初体验", Desc: "生成第 1 份工作报告", Category: CategoryExplorer, Icon: "monitor", Tier: TierBronze, Target: 1},
	{Code: "report_5", Name: "总结小能手", Desc: "生成 5 份工作报告", Category: CategoryExplorer, Icon: "monitor", Tier: TierBronze, Target: 5},
	{Code: "report_20", Name: "文案收割机", Desc: "生成 20 份工作报告", Category: CategoryExplorer, Icon: "monitor", Tier: TierSilver, Target: 20},
	{Code: "report_ai_10", Name: "AI 指挥家", Desc: "生成 10 份 AI 报告", Category: CategoryExplorer, Icon: "monitor", Tier: TierGold, Target: 10},
	{Code: "report_team_1", Name: "团队分析师", Desc: "生成第 1 份团队报告", Category: CategoryExplorer, Icon: "bars", Tier: TierBronze, Target: 1},
	{Code: "report_group_1", Name: "工作组观察员", Desc: "生成第 1 份工作组报告", Category: CategoryExplorer, Icon: "bars", Tier: TierBronze, Target: 1},
	{Code: "report_week_1", Name: "周报打工人", Desc: "生成第 1 份周报", Category: CategoryExplorer, Icon: "clock", Tier: TierBronze, Target: 1},
	{Code: "report_month_1", Name: "月度复盘家", Desc: "生成第 1 份月报", Category: CategoryExplorer, Icon: "clock", Tier: TierSilver, Target: 1},

	// ───────────── 八、个性与陪伴（17）─────────────
	{Code: "avatar_set", Name: "真容示人", Desc: "设置个人头像", Category: CategoryPersonality, Icon: "user", Tier: TierBronze, Target: 1},
	{Code: "bg_set", Name: "装修大师", Desc: "设置平台背景图", Category: CategoryPersonality, Icon: "view", Tier: TierBronze, Target: 1},
	{Code: "tag_1", Name: "标签开辟者", Desc: "创建第 1 个标签", Category: CategoryPersonality, Icon: "tag", Tier: TierBronze, Target: 1},
	{Code: "tag_10", Name: "分类大师", Desc: "创建 10 个标签", Category: CategoryPersonality, Icon: "tag", Tier: TierSilver, Target: 10},
	{Code: "days_1", Name: "初来乍到", Desc: "加入平台第 1 天", Category: CategoryPersonality, Icon: "sun", Tier: TierBronze, Target: 1},
	{Code: "days_7", Name: "一周养成", Desc: "陪伴平台满 7 天", Category: CategoryPersonality, Icon: "sun", Tier: TierBronze, Target: 7},
	{Code: "days_30", Name: "月度之约", Desc: "陪伴平台满 30 天", Category: CategoryPersonality, Icon: "sun", Tier: TierSilver, Target: 30},
	{Code: "days_100", Name: "百日之约", Desc: "陪伴平台满 100 天", Category: CategoryPersonality, Icon: "star", Tier: TierGold, Target: 100},
	{Code: "days_365", Name: "周年之约", Desc: "陪伴平台满 365 天", Category: CategoryPersonality, Icon: "crown", Tier: TierPlatinum, Target: 365},
	{Code: "active_3", Name: "初露锋芒", Desc: "累计 3 天有操作记录", Category: CategoryPersonality, Icon: "check", Tier: TierBronze, Target: 3},
	{Code: "active_7", Name: "七日之约", Desc: "累计 7 天有操作记录", Category: CategoryPersonality, Icon: "check", Tier: TierSilver, Target: 7},
	{Code: "active_30", Name: "活跃常青树", Desc: "累计 30 天有操作记录", Category: CategoryPersonality, Icon: "check", Tier: TierGold, Target: 30},
	{Code: "early_5", Name: "早鸟达人", Desc: "清晨 5-8 点创建 5 个任务", Category: CategoryPersonality, Icon: "sun", Tier: TierSilver, Target: 5},
	{Code: "late_complete_1", Name: "深夜加班人", Desc: "凌晨 0-5 点完成过任务", Category: CategoryPersonality, Icon: "moon", Tier: TierBronze, Target: 1, IsMystery: true},
	{Code: "weekend_10", Name: "周末战士", Desc: "周末完成 10 个任务", Category: CategoryPersonality, Icon: "star", Tier: TierSilver, Target: 10},
	{Code: "all_rounder", Name: "全能选手", Desc: "创建、完成、聊天、提问题各至少 1 次", Category: CategoryPersonality, Icon: "crown", Tier: TierGold, Target: 1},
	{Code: "veteran", Name: "传奇人物", Desc: "解锁 50 个其他成就", Category: CategoryPersonality, Icon: "crown", Tier: TierPlatinum, Target: 50},
}

// Count 返回成就总数（应为 100）
func Count() int { return len(defList) }
