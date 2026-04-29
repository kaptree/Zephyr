<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getUsers, createUser, updateUser, deleteUser, getDepartments } from '@/services/admin'
import type { User, Department } from '@/types'

interface UserRow {
  id: string
  username: string
  name: string
  dept_name: string
  dept_id: string
  role: string
  rank: string
  phone: string
  email: string
  is_active: boolean
}

const users = ref<UserRow[]>([])
const loading = ref(false)
const loadError = ref('')

const departments = ref<Department[]>([])
const allDepts = ref<{ id: string; name: string }[]>([])

const showModal = ref(false)
const editingUserId = ref<string | null>(null)
const formUsername = ref('')
const formName = ref('')
const formPassword = ref('')
const formRole = ref('user')
const formRank = ref('')
const formPhone = ref('')
const formEmail = ref('')
const formDeptId = ref('')
const formIsActive = ref(true)
const submitting = ref(false)
const formError = ref('')

const roleMap: Record<string, string> = {
  super_admin: '系统管理员',
  dept_admin: '部门管理员',
  group_leader: '组长',
  user: '普通民警',
  screen_role: '大屏角色',
}

function flattenDepts(depts: Department[], prefix = ''): void {
  for (const d of depts) {
    allDepts.value.push({ id: d.id, name: prefix + d.name })
    if (d.children) flattenDepts(d.children, prefix + '  ')
  }
}

async function loadData() {
  loading.value = true
  loadError.value = ''
  try {
    const [deptRes, userRes] = await Promise.all([
      getDepartments(false),
      getUsers({ page: 1, page_size: 100 }),
    ])
    departments.value = deptRes.data as unknown as Department[]
    allDepts.value = []
    flattenDepts(departments.value)
    const rawData = (userRes.data as unknown as { data: any[] }).data || []
    users.value = rawData.map((u: any) => ({
      ...u,
      dept_name: u.department?.name || u.dept_name || '未分配',
      dept_id: u.dept_id || u.department?.id || '',
      role: u.role || 'user',
      rank: u.rank || '',
      phone: u.phone || '',
      email: u.email || '',
    }))
  } catch {
    loadError.value = '加载人员列表失败'
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

function openCreate() {
  editingUserId.value = null
  formUsername.value = ''
  formName.value = ''
  formPassword.value = ''
  formRole.value = 'user'
  formRank.value = ''
  formPhone.value = ''
  formEmail.value = ''
  formDeptId.value = ''
  formIsActive.value = true
  formError.value = ''
  showModal.value = true
}

function openEdit(user: UserRow) {
  editingUserId.value = user.id
  formUsername.value = user.username
  formName.value = user.name
  formPassword.value = ''
  formRole.value = user.role
  formRank.value = user.rank || ''
  formPhone.value = user.phone || ''
  formEmail.value = user.email || ''
  formDeptId.value = (user.dept_id || '').trim()
  formIsActive.value = user.is_active !== false
  formError.value = ''
  showModal.value = true
}

async function handleSubmit() {
  if (!formName.value.trim()) {
    formError.value = '请输入姓名'
    return
  }
  if (!editingUserId.value && !formUsername.value.trim()) {
    formError.value = '请输入用户名'
    return
  }
  if (!editingUserId.value && !formPassword.value.trim()) {
    formError.value = '请输入密码'
    return
  }

  const deptId = (formDeptId.value || '').trim()

  submitting.value = true
  formError.value = ''
  try {
    if (editingUserId.value) {
      await updateUser(editingUserId.value, {
        name: formName.value.trim(),
        role: formRole.value,
        rank: (formRank.value || '').trim(),
        phone: (formPhone.value || '').trim(),
        email: (formEmail.value || '').trim(),
        dept_id: deptId,
        is_active: formIsActive.value,
      })
    } else {
      await createUser({
        username: formUsername.value.trim(),
        name: formName.value.trim(),
        password: formPassword.value,
        role: formRole.value,
        rank: (formRank.value || '').trim(),
        phone: (formPhone.value || '').trim(),
        email: (formEmail.value || '').trim(),
        dept_id: deptId,
      })
    }
    showModal.value = false
    await loadData()
  } catch (e: unknown) {
    const err = e as { response?: { data?: { message?: string } } }
    formError.value = err?.response?.data?.message || '操作失败'
  } finally {
    submitting.value = false
  }
}

async function handleDelete(user: UserRow) {
  if (!confirm(`确定要删除人员"${user.name}"吗？`)) return
  try {
    await deleteUser(user.id)
    await loadData()
  } catch (e: unknown) {
    const err = e as { response?: { data?: { message?: string } } }
    alert(err?.response?.data?.message || '删除失败')
  }
}

function getRoleClass(role: string) {
  const map: Record<string, string> = {
    super_admin: 'bg-amber-100 text-amber-700',
    dept_admin: 'bg-blue-100 text-blue-700',
    group_leader: 'bg-green-100 text-green-700',
    user: 'bg-slate-100 text-slate-600',
  }
  return map[role] || 'bg-slate-100 text-slate-600'
}

function getStatusClass(active: boolean) {
  return active ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h2 class="text-lg font-semibold text-slate-900">人员库管理</h2>
      <button class="btn-primary text-sm" @click="openCreate">新建人员</button>
    </div>

    <div class="bg-white rounded-card border border-slate-100 overflow-hidden">
      <div v-if="loading" class="p-8 text-center text-sm text-slate-400">加载中...</div>
      <div v-else-if="loadError" class="p-8 text-center text-sm text-red-400">
        {{ loadError }}
        <button class="block mx-auto mt-2 text-xs text-blue-500 hover:underline" @click="loadData">重试</button>
      </div>
      <table v-else class="w-full">
        <thead>
          <tr class="border-b border-slate-100 bg-slate-50">
            <th class="text-left px-4 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">姓名</th>
            <th class="text-left px-4 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">用户名</th>
            <th class="text-left px-4 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">部门</th>
            <th class="text-left px-4 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">角色</th>
            <th class="text-left px-4 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">状态</th>
            <th class="text-right px-4 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-50">
          <tr v-for="user in users" :key="user.id" class="hover:bg-slate-50/50 transition-smooth">
            <td class="px-4 py-3">
              <div class="flex items-center gap-3">
                <div class="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center text-xs font-medium text-blue-600">
                  {{ user.name.charAt(0) }}
                </div>
                <div>
                  <div class="text-sm font-medium text-slate-900">{{ user.name }}</div>
                  <div class="text-xs text-slate-400" v-if="user.rank">{{ user.rank }}</div>
                </div>
              </div>
            </td>
            <td class="px-4 py-3 text-sm text-slate-600 font-mono">{{ user.username }}</td>
            <td class="px-4 py-3 text-sm text-slate-600">{{ user.dept_name }}</td>
            <td class="px-4 py-3">
              <span :class="['text-xs px-2 py-0.5 rounded-tag font-medium', getRoleClass(user.role)]">
                {{ roleMap[user.role] || user.role }}
              </span>
            </td>
            <td class="px-4 py-3">
              <span :class="['text-xs px-2 py-0.5 rounded-tag font-medium', getStatusClass(user.is_active !== false)]">
                {{ user.is_active !== false ? '正常' : '禁用' }}
              </span>
            </td>
            <td class="px-4 py-3 text-right">
              <div class="flex items-center justify-end gap-1">
                <button class="text-xs px-2 py-1 bg-blue-50 text-blue-600 rounded hover:bg-blue-100 transition-smooth" @click="openEdit(user)">编辑</button>
                <button class="text-xs px-2 py-1 bg-red-50 text-red-600 rounded hover:bg-red-100 transition-smooth" @click="handleDelete(user)">删除</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 新建/编辑人员模态框 -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-start justify-center pt-[8vh]">
        <div class="overlay-backdrop" @click="showModal = false" />
        <div class="relative z-50 bg-white rounded-card shadow-modal w-full max-w-lg mx-4 p-6 animate-fade-in">
          <h3 class="text-base font-semibold text-slate-900 mb-4">{{ editingUserId ? '编辑人员' : '新建人员' }}</h3>

          <form @submit.prevent="handleSubmit" class="space-y-3 max-h-[70vh] overflow-y-auto pr-1 scrollbar-thin">
            <div class="grid grid-cols-2 gap-3">
              <div>
                <span class="text-xs text-slate-500 mb-1 block">用户名</span>
                <input v-model="formUsername" class="input-field !py-1.5 !text-sm" placeholder="登录用户名" :disabled="!!editingUserId" />
              </div>
              <div>
                <span class="text-xs text-slate-500 mb-1 block">密码</span>
                <input v-model="formPassword" type="password" class="input-field !py-1.5 !text-sm" :placeholder="editingUserId ? '留空不修改' : '设置密码'" />
              </div>
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div>
                <span class="text-xs text-slate-500 mb-1 block">姓名</span>
                <input v-model="formName" class="input-field !py-1.5 !text-sm" placeholder="真实姓名" />
              </div>
              <div>
                <span class="text-xs text-slate-500 mb-1 block">警衔/职级</span>
                <input v-model="formRank" class="input-field !py-1.5 !text-sm" placeholder="如: 二级警督" />
              </div>
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div>
                <span class="text-xs text-slate-500 mb-1 block">角色</span>
                <select v-model="formRole" class="input-field !py-1.5 !text-sm">
                  <option value="super_admin">系统管理员</option>
                  <option value="dept_admin">部门管理员</option>
                  <option value="group_leader">组长</option>
                  <option value="user">普通民警</option>
                  <option value="screen_role">大屏角色</option>
                </select>
              </div>
              <div>
                <span class="text-xs text-slate-500 mb-1 block">所属部门</span>
                <select v-model="formDeptId" class="input-field !py-1.5 !text-sm">
                  <option value="">未分配</option>
                  <option v-for="d in allDepts" :key="d.id" :value="d.id">{{ d.name }}</option>
                </select>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div>
                <span class="text-xs text-slate-500 mb-1 block">手机号</span>
                <input v-model="formPhone" class="input-field !py-1.5 !text-sm" placeholder="手机号" />
              </div>
              <div>
                <span class="text-xs text-slate-500 mb-1 block">邮箱</span>
                <input v-model="formEmail" class="input-field !py-1.5 !text-sm" placeholder="邮箱地址" />
              </div>
            </div>

            <div v-if="editingUserId" class="flex items-center gap-2">
              <span class="text-xs text-slate-500">账号状态</span>
              <label class="flex items-center gap-1.5 cursor-pointer">
                <input v-model="formIsActive" type="checkbox" class="w-4 h-4 text-blue-500 rounded" />
                <span class="text-xs text-slate-700">{{ formIsActive ? '正常' : '禁用' }}</span>
              </label>
            </div>

            <p v-if="formError" class="text-xs text-red-500 bg-red-50 px-3 py-2 rounded-btn">{{ formError }}</p>

            <div class="flex justify-end gap-3 pt-4 border-t border-slate-100">
              <button type="button" class="btn-secondary text-xs !py-1.5 !px-4" @click="showModal = false">取消</button>
              <button type="submit" class="btn-primary text-xs !py-1.5 !px-4" :disabled="submitting">
                {{ submitting ? '提交中...' : editingUserId ? '保存修改' : '创建人员' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>
  </div>
</template>
