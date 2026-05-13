package main

import (
	"fmt"
	"log"
	"time"

	"labelpro-server/internal/config"
	"labelpro-server/internal/database"
	"labelpro-server/internal/logger"
	"labelpro-server/internal/models"
	"labelpro-server/internal/utils"

	"github.com/google/uuid"
)

func main() {
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	if err := logger.Init(
		cfg.Log.Level,
		cfg.Log.Format,
		cfg.Log.OutputDir,
		cfg.Log.MaxSizeMB,
		cfg.Log.MaxBackups,
		cfg.Log.MaxAgeDays,
		cfg.Log.Compress,
		cfg.Log.EnableConsole,
	); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logger.Sync()

	if err := database.InitPostgres(cfg); err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	if err := database.DB.AutoMigrate(
		&models.User{},
		&models.Department{},
		&models.RolePermission{},
		&models.Note{},
		&models.Tag{},
		&models.Template{},
		&models.NoteAssignee{},
		&models.NoteAttachment{},
		&models.WorkGroup{},
		&models.WorkGroupMember{},
		&models.PresetGroup{},
		&models.PresetGroupMember{},
		&models.CollaborationRoom{},
		&models.Reminder{},
		&models.LedgerEntry{},
	); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	fmt.Println("数据库迁移完成")

	seedCount := seedAll()
	fmt.Printf("\n测试数据初始化完成! 共创建: %d 条记录\n", seedCount)
}

func seedAll() int {
	total := 0
	total += seedPermissions()
	total += seedDepartments()
	total += seedUsers()
	total += seedTags()
	total += seedTemplates()
	total += seedNotes()
	return total
}

func seedPermissions() int {
	var count int64
	database.DB.Model(&models.RolePermission{}).Count(&count)
	if count > 0 {
		fmt.Println("权限数据已存在，跳过")
		return 0
	}

	perms := []models.RolePermission{
		{Role: "super_admin", Resource: "note", Action: "create", Scope: "global"},
		{Role: "super_admin", Resource: "note", Action: "read", Scope: "global"},
		{Role: "super_admin", Resource: "note", Action: "update", Scope: "global"},
		{Role: "super_admin", Resource: "note", Action: "delete", Scope: "global"},
		{Role: "super_admin", Resource: "note", Action: "remind", Scope: "global"},
		{Role: "super_admin", Resource: "user", Action: "create", Scope: "global"},
		{Role: "super_admin", Resource: "user", Action: "read", Scope: "global"},
		{Role: "super_admin", Resource: "user", Action: "update", Scope: "global"},
		{Role: "super_admin", Resource: "user", Action: "delete", Scope: "global"},
		{Role: "super_admin", Resource: "department", Action: "manage", Scope: "global"},
		{Role: "super_admin", Resource: "tag", Action: "manage", Scope: "global"},
		{Role: "super_admin", Resource: "template", Action: "manage", Scope: "global"},
		{Role: "dept_admin", Resource: "note", Action: "create", Scope: "department"},
		{Role: "dept_admin", Resource: "note", Action: "read", Scope: "department"},
		{Role: "dept_admin", Resource: "note", Action: "update", Scope: "department"},
		{Role: "dept_admin", Resource: "note", Action: "remind", Scope: "department"},
		{Role: "dept_admin", Resource: "user", Action: "read", Scope: "department"},
		{Role: "dept_admin", Resource: "user", Action: "update", Scope: "department"},
		{Role: "group_leader", Resource: "note", Action: "create", Scope: "group"},
		{Role: "group_leader", Resource: "note", Action: "read", Scope: "group"},
		{Role: "group_leader", Resource: "note", Action: "update", Scope: "group"},
		{Role: "group_leader", Resource: "note", Action: "remind", Scope: "group"},
		{Role: "member", Resource: "note", Action: "create", Scope: "self"},
		{Role: "member", Resource: "note", Action: "read", Scope: "self"},
	}
	database.DB.Create(&perms)
	fmt.Printf("  ✓ 权限矩阵: %d 条\n", len(perms))
	return len(perms)
}

func seedDepartments() int {
	var count int64
	database.DB.Model(&models.Department{}).Count(&count)
	if count > 0 {
		fmt.Println("部门数据已存在，跳过")
		return 0
	}

	gaID := uuid.New()
	xjID := uuid.New()
	zaID := uuid.New()
	waID := uuid.New()
	jszdID := uuid.New()
	qbzdID := uuid.New()
	yjzdID := uuid.New()
	zzzdID := uuid.New()

	depts := []models.Department{
		{ID: gaID, Name: "市公司局", ParentID: nil, Level: 1, SortOrder: 1},
		{ID: xjID, Name: "刑警支队", ParentID: &gaID, Level: 2, SortOrder: 1},
		{ID: zaID, Name: "治安支队", ParentID: &gaID, Level: 2, SortOrder: 2},
		{ID: waID, Name: "网安支队", ParentID: &gaID, Level: 2, SortOrder: 3},
		{ID: jszdID, Name: "技术侦查中队", ParentID: &xjID, Level: 3, SortOrder: 1},
		{ID: qbzdID, Name: "情报研判中队", ParentID: &xjID, Level: 3, SortOrder: 2},
		{ID: yjzdID, Name: "应急管理大队", ParentID: &zaID, Level: 3, SortOrder: 1},
		{ID: zzzdID, Name: "作战指挥中队", ParentID: &waID, Level: 3, SortOrder: 1},
	}
	database.DB.Create(&depts)
	fmt.Printf("  ✓ 部门: %d 个\n", len(depts))
	return len(depts)
}

func seedUsers() int {
	var count int64
	database.DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		fmt.Println("用户数据已存在，跳过")
		return 0
	}

	var ga, xj, za, wa, jszd, qbzd, yjzd, zzzd models.Department
	database.DB.Where("name = ?", "市公司局").First(&ga)
	database.DB.Where("name = ?", "刑警支队").First(&xj)
	database.DB.Where("name = ?", "治安支队").First(&za)
	database.DB.Where("name = ?", "网安支队").First(&wa)
	database.DB.Where("name = ?", "技术侦查中队").First(&jszd)
	database.DB.Where("name = ?", "情报研判中队").First(&qbzd)
	database.DB.Where("name = ?", "应急管理大队").First(&yjzd)
	database.DB.Where("name = ?", "作战指挥中队").First(&zzzd)

	pwd, _ := utils.HashPassword("Admin@123")

	users := []models.User{
		{ID: uuid.New(), Username: "admin", Name: "张局长", DepartmentID: &ga.ID, Role: "super_admin", Rank: "一级警监", Phone: "13800001001", Email: "admin@police.cn", PasswordHash: pwd, IsActive: true},
		{ID: uuid.New(), Username: "wang", Name: "王大队", DepartmentID: &xj.ID, Role: "dept_admin", Rank: "二级警督", Phone: "13800001002", Email: "wang@police.cn", PasswordHash: pwd, IsActive: true},
		{ID: uuid.New(), Username: "li", Name: "李中队", DepartmentID: &jszd.ID, Role: "group_leader", Rank: "三级警督", Phone: "13800001003", Email: "li@police.cn", PasswordHash: pwd, IsActive: true},
		{ID: uuid.New(), Username: "zhang", Name: "张三", DepartmentID: &jszd.ID, Role: "member", Rank: "一级警员", Phone: "13800001004", Email: "zhang@police.cn", PasswordHash: pwd, IsActive: true},
		{ID: uuid.New(), Username: "zhao", Name: "赵六", DepartmentID: &qbzd.ID, Role: "member", Rank: "二级警员", Phone: "13800001005", Email: "zhao@police.cn", PasswordHash: pwd, IsActive: true},
		{ID: uuid.New(), Username: "sun", Name: "孙队", DepartmentID: &za.ID, Role: "dept_admin", Rank: "二级警督", Phone: "13800001006", Email: "sun@police.cn", PasswordHash: pwd, IsActive: true},
		{ID: uuid.New(), Username: "zhou", Name: "周干事", DepartmentID: &yjzd.ID, Role: "member", Rank: "一级警员", Phone: "13800001007", Email: "zhou@police.cn", PasswordHash: pwd, IsActive: true},
		{ID: uuid.New(), Username: "wu", Name: "吴主任", DepartmentID: &wa.ID, Role: "dept_admin", Rank: "二级警督", Phone: "13800001008", Email: "wu@police.cn", PasswordHash: pwd, IsActive: true},
		{ID: uuid.New(), Username: "chen", Name: "陈班长", DepartmentID: &zzzd.ID, Role: "group_leader", Rank: "三级警督", Phone: "13800001009", Email: "chen@police.cn", PasswordHash: pwd, IsActive: true},
		{ID: uuid.New(), Username: "liu", Name: "刘探员", DepartmentID: &jszd.ID, Role: "member", Rank: "二级警员", Phone: "13800001010", Email: "liu@police.cn", PasswordHash: pwd, IsActive: true},
	}
	database.DB.Create(&users)
	fmt.Printf("  ✓ 用户: %d 个 (密码均为 Admin@123)\n", len(users))
	return len(users)
}

func seedTags() int {
	var count int64
	database.DB.Model(&models.Tag{}).Count(&count)
	if count > 0 {
		fmt.Println("标签数据已存在，跳过")
		return 0
	}

	total := 0

	priorityTags := []models.Tag{
		{Name: "紧急", Color: "#EF4444", Category: "优先级", Scope: "system", SortOrder: 1},
		{Name: "重要", Color: "#F59E0B", Category: "优先级", Scope: "system", SortOrder: 2},
		{Name: "普通", Color: "#3B82F6", Category: "优先级", Scope: "system", SortOrder: 3},
	}
	database.DB.Create(&priorityTags)
	total += len(priorityTags)

	categories := []struct {
		Name     string
		Color    string
		Children []string
	}{
		{"治安管控", "#2563EB", []string{"行业检查", "场所巡查", "隐患排查", "治安巡逻", "重点人员管控"}},
		{"巡逻防控", "#059669", []string{"街面巡逻", "卡点盘查", "视频巡查", "便衣巡逻", "武装联勤"}},
		{"矛盾调解", "#7C3AED", []string{"邻里纠纷", "家庭矛盾", "物业纠纷", "劳资纠纷", "征地拆迁"}},
		{"线索核查", "#EA580C", []string{"刑事线索", "治安线索", "网安线索", "禁毒线索", "经侦线索"}},
		{"专案办理", "#DC2626", []string{"刑事专案", "治安专案", "毒品专案", "经侦专案", "网安专案"}},
		{"安保勤务", "#CA8A04", []string{"大型活动安保", "警卫任务", "安全检查", "应急处置", "维稳处突"}},
		{"舆情处置", "#0891B2", []string{"舆情监测", "舆论引导", "信息核查", "辟谣发布", "网络巡查"}},
		{"内勤事务", "#4B5563", []string{"档案管理", "装备管理", "会议保障", "报表统计", "材料报送"}},
		{"其他工作", "#6B7280", []string{"培训学习", "社区走访", "协助外单位", "临时勤务", "其他"}},
	}

	for i, cat := range categories {
		parent := models.Tag{
			Name:     cat.Name,
			Color:    cat.Color,
			Category: "一级分类",
			Scope:    "system",
			SortOrder: i + 10,
		}
		database.DB.Create(&parent)
		total++

		for j, childName := range cat.Children {
			child := models.Tag{
				Name:     childName,
				Color:    cat.Color,
				Category: "二级分类",
				Scope:    "system",
				ParentID: &parent.ID,
				SortOrder: j + 1,
			}
			database.DB.Create(&child)
			total++
		}
	}

	fmt.Printf("  ✓ 标签: %d 个 (含 %d 个一级分类, 各含 5 个子标签)\n", total, len(categories))
	return total
}

func seedTemplates() int {
	var count int64
	database.DB.Model(&models.Template{}).Count(&count)
	if count > 0 {
		fmt.Println("模板数据已存在，跳过")
		return 0
	}

	templates := []models.Template{
		{
			Name: "数据分析研判模板", Type: "data_analysis", IsSystem: true, Layout: "2",
			Fields: `[{"name":"数据来源","type":"text","required":true,"order":1},{"name":"分析周期","type":"date","required":true,"order":2},{"name":"分析结论","type":"rich-text","required":true,"order":3},{"name":"处置建议","type":"textarea","required":false,"order":4}]`,
		},
		{
			Name: "专项行动方案模板", Type: "special_project", IsSystem: true, Layout: "4",
			Fields: `[{"name":"行动名称","type":"text","required":true,"order":1},{"name":"行动时间","type":"date","required":true,"order":2},{"name":"参与单位","type":"multi-select","required":true,"order":3},{"name":"任务分工","type":"rich-text","required":true,"order":4}]`,
		},
		{
			Name: "紧急协查通报模板", Type: "emergency_canvas", IsSystem: true, Layout: "1",
			Fields: `[{"name":"案由","type":"text","required":true,"order":1},{"name":"协查对象","type":"textarea","required":true,"order":2},{"name":"协查要求","type":"rich-text","required":true,"order":3},{"name":"反馈时限","type":"date","required":true,"order":4}]`,
		},
		{
			Name: "协同作战方案模板", Type: "collaborative_writing", IsSystem: true, Layout: "6",
			Fields: `[{"name":"作战目标","type":"text","required":true,"order":1},{"name":"参战力量","type":"textarea","required":true,"order":2},{"name":"部署方案","type":"rich-text","required":true,"order":3}]`,
		},
		{
			Name: "日常工作任务模板", Type: "default", IsSystem: true, Layout: "1",
			Fields: `[{"name":"任务描述","type":"textarea","required":true,"order":1},{"name":"完成标准","type":"textarea","required":false,"order":2}]`,
		},
	}
	database.DB.Create(&templates)
	fmt.Printf("  ✓ 模板: %d 个\n", len(templates))
	return len(templates)
}

func seedNotes() int {
	var count int64
	database.DB.Model(&models.Note{}).Count(&count)
	if count > 0 {
		fmt.Println("任务数据已存在，跳过")
		return 0
	}

	var users []models.User
	database.DB.Find(&users)
	userMap := make(map[string]models.User)
	for _, u := range users {
		userMap[u.Username] = u
	}

	var tags []models.Tag
	database.DB.Find(&tags)
	tagMap := make(map[string]models.Tag)
	for _, t := range tags {
		tagMap[t.Name] = t
	}

	var depts []models.Department
	database.DB.Find(&depts)
	deptMap := make(map[string]models.Department)
	for _, d := range depts {
		deptMap[d.Name] = d
	}

	noteCount := 0

	due24h := time.Now().Add(24 * time.Hour)
	n1 := createNote(userMap["zhang"], deptMap["技术侦查中队"], []models.Tag{tagMap["重要"], tagMap["刑事线索"]},
		"嫌疑人活动轨迹分析", "<p>针对近期连环盗窃案嫌疑人张某的活动轨迹进行分析研判，需梳理其近14天的活动规律、落脚点和同案人员关系网。</p>",
		"yellow", "self", &due24h)
	_ = n1
	noteCount++

	due6h := time.Now().Add(6 * time.Hour)
	n2 := createNote(userMap["li"], deptMap["技术侦查中队"], []models.Tag{tagMap["紧急"], tagMap["刑事线索"]},
		"紧急协查：涉黑团伙骨干在逃", "<p>上级通报涉黑团伙骨干成员李某可能已潜入我市，需立即组织警力开展摸排布控，务必将嫌疑人尽快抓捕归案。</p>",
		"red", "assigned", &due6h)
	addAssignee(n2.ID, userMap["li"].ID, userMap["zhang"].ID)
	_ = n2
	noteCount++

	due48h := time.Now().Add(48 * time.Hour)
	n3 := createNote(userMap["admin"], deptMap["市公司局"], []models.Tag{tagMap["紧急"], tagMap["刑事专案"]},
		"「雷霆2026」夏季治安打击专项行动", "<p>根据公司部和省厅统一部署，在全市范围内开展「雷霆2026」夏季治安打击专项行动。主要目标：严厉打击涉黑涉恶、黄赌毒、电信诈骗等突出违法犯罪。</p>",
		"red", "assigned", &due48h)
	addAssignee(n3.ID, userMap["admin"].ID, userMap["wang"].ID)
	addAssignee(n3.ID, userMap["admin"].ID, userMap["sun"].ID)
	addAssignee(n3.ID, userMap["admin"].ID, userMap["wu"].ID)
	_ = n3
	noteCount++

	due72h := time.Now().Add(72 * time.Hour)
	createNote(userMap["zhao"], deptMap["情报研判中队"], []models.Tag{tagMap["普通"], tagMap["场所巡查"]},
		"重点场所人员信息采集任务", "<p>完成辖区内旅馆、网吧、KTV等重点场所的外来人员信息采集，确保信息准确率达100%，本周五前完成汇总上报。</p>",
		"yellow", "self", &due72h)
	noteCount++

	due120h := time.Now().Add(120 * time.Hour)
	createNote(userMap["sun"], deptMap["治安支队"], []models.Tag{tagMap["重要"], tagMap["治安巡逻"]},
		"端午节期间重点区域巡控布防方案", "<p>制定端午节期间火车站、商业圈、旅游景点等重点区域巡控布防方案，合理调配警力资源。</p>",
		"yellow", "self", &due120h)
	noteCount++

	now := time.Now()
	n6 := createNote(userMap["zhou"], deptMap["应急管理大队"], []models.Tag{tagMap["普通"], tagMap["应急处置"]},
		"应急指挥系统链路调试", "<p>配合技术部门完成应急指挥系统的链路调试，确保会议中心、各分局、车载终端的音视频正常联通。</p>",
		"green", "self", nil)
	n6.ColorStatus = "green"
	n6.IsArchived = true
	n6.ArchiveTime = &now
	n6.CompletedAt = &now
	database.DB.Save(n6)
	_ = n6
	noteCount++

	due36h := time.Now().Add(36 * time.Hour)
	createNote(userMap["wu"], deptMap["网安支队"], []models.Tag{tagMap["重要"], tagMap["应急处置"]},
		"涉网案件电子取证分析", "<p>对近期侦办的网络诈骗案件进行电子数据取证分析，包括服务器日志分析、电子支付记录追踪和通讯记录恢复。</p>",
		"yellow", "self", &due36h)
	noteCount++

	due2h := time.Now().Add(2 * time.Hour)
	n8 := createNote(userMap["chen"], deptMap["作战指挥中队"], []models.Tag{tagMap["紧急"], tagMap["刑事线索"]},
		"重大警情：网络攻击溯源", "<p>市政务云平台遭受DDoS攻击，需立即开展溯源分析，封堵攻击IP，评估数据泄露风险，并形成技术通报。</p>",
		"red", "assigned", &due2h)
	addAssignee(n8.ID, userMap["chen"].ID, userMap["wu"].ID)
	_ = n8
	noteCount++

	n9 := createNote(userMap["wang"], deptMap["刑警支队"], []models.Tag{tagMap["重要"], tagMap["会议保障"]},
		"全市刑侦工作月度分析会议纪要", "<p>总结上月刑侦工作成效，分析当前刑事犯罪形势，部署本月重点工作。重点议题：电诈打防、命案积案攻坚、追逃工作。</p>",
		"green", "self", nil)
	n9.ColorStatus = "green"
	n9.IsArchived = true
	n9.ArchiveTime = &now
	n9.CompletedAt = &now
	database.DB.Save(n9)
	_ = n9
	noteCount++

	due8h := time.Now().Add(8 * time.Hour)
	createNote(userMap["liu"], deptMap["技术侦查中队"], []models.Tag{tagMap["普通"], tagMap["其他"]},
		"嫌疑车辆轨迹查询", "<p>查询车牌号京A·XXXXX在2026年4月20日至25日期间的行驶轨迹，包括ETC记录、卡口抓拍、停车场进出记录。</p>",
		"yellow", "self", &due8h)
	noteCount++

	dueCollab := time.Now().Add(24 * time.Hour)
	c1 := createNote(userMap["wang"], deptMap["刑警支队"], []models.Tag{tagMap["重要"], tagMap["刑事线索"]},
		"协同研判：跨区域系列盗窃案件串并分析", "<p>联合治安支队、网安支队共同研判近期城区系列盗窃案件，分析案件特征、作案手法，进行案件串并，制定协同作战方案。</p>",
		"blue", "collaboration", &dueCollab)
	_ = c1
	noteCount++

	createNote(userMap["sun"], deptMap["治安支队"], []models.Tag{tagMap["重要"], tagMap["隐患排查"]},
		"协同排查：重点场所联合安全检查行动", "<p>联合消防、市场监管等部门，对辖区内重点场所开展联合安全检查，各成员单位分工协作，汇总检查结果并制定整改方案。</p>",
		"blue", "collaboration", &dueCollab)
	noteCount++

	fmt.Printf("  ✓ 任务: %d 条 (2条已归档, 3条紧急盯办, 2条协同协作)\n", noteCount)
	return noteCount
}

func createNote(owner models.User, dept models.Department, tags []models.Tag, title, content, colorStatus, sourceType string, dueTime *time.Time) *models.Note {
	note := &models.Note{
		ID:           uuid.New(),
		Title:        title,
		Content:      content,
		ColorStatus:  colorStatus,
		SourceType:   sourceType,
		TemplateType: "default",
		CreatorID:    owner.ID,
		OwnerID:      owner.ID,
		DepartmentID: &dept.ID,
		DueTime:      dueTime,
		Tags:         tags,
	}

	year := time.Now().Year()
	var maxSeq int64
	database.DB.Model(&models.Note{}).
		Where("serial_no LIKE ?", fmt.Sprintf("资警轻燕〔%d〕%%", year)).
		Count(&maxSeq)
	note.SerialNo = fmt.Sprintf("资警轻燕〔%d〕%04d号", year, int(maxSeq)+1)

	database.DB.Create(note)

	assignee := models.NoteAssignee{
		NoteID:     note.ID,
		UserID:     owner.ID,
		RoleInNote: "initiator",
	}
	database.DB.Create(&assignee)

	entry := models.LedgerEntry{
		NoteID:       note.ID,
		UserID:       owner.ID,
		Action:       "create",
		ActionDetail: "创建任务",
	}
	database.DB.Create(&entry)

	return note
}

func addAssignee(noteID, initiatorID, userID uuid.UUID) {
	assignee := models.NoteAssignee{
		NoteID:     noteID,
		UserID:     userID,
		RoleInNote: "member",
	}
	database.DB.Create(&assignee)
}

var _ = utils.HashPassword
