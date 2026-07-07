<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex-1 sm:max-w-80">
            <input
              v-model="searchQuery"
              type="text"
              class="input"
              :placeholder="t('admin.ipBans.searchPlaceholder')"
              @input="handleSearch"
            />
          </div>
          <Select
            v-model="filters.status"
            :options="filterStatusOptions"
            class="w-40"
            @change="reloadFromFirstPage"
          />
          <div class="flex flex-1 flex-wrap items-center justify-end gap-2">
            <button
              class="btn btn-secondary"
              :disabled="loading"
              :title="t('common.refresh')"
              @click="loadBans"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
            <button class="btn btn-primary" @click="openCreateDialog">
              <Icon name="plus" size="md" class="mr-1" />
              {{ t('admin.ipBans.createRule') }}
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable
          :columns="columns"
          :data="bans"
          :loading="loading"
          :server-side-sort="true"
          default-sort-key="created_at"
          default-sort-order="desc"
          @sort="handleSort"
        >
          <template #cell-pattern="{ value }">
            <code class="rounded bg-gray-100 px-2 py-1 font-mono text-sm text-gray-900 dark:bg-dark-700 dark:text-gray-100">
              {{ value }}
            </code>
          </template>

          <template #cell-status="{ row }">
            <div class="flex flex-col gap-1">
              <span :class="['badge', getStatusClass(row)]">
                {{ getStatusLabel(row) }}
              </span>
              <span v-if="isExpired(row)" class="text-xs text-red-500">
                {{ t('admin.ipBans.expired') }}
              </span>
            </div>
          </template>

          <template #cell-reason="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">
              {{ value || '-' }}
            </span>
          </template>

          <template #cell-hit_count="{ value }">
            <span class="font-mono text-sm text-gray-700 dark:text-gray-300">
              {{ formatCompactNumber(value) }}
            </span>
          </template>

          <template #cell-last_hit_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">
              {{ value ? formatDateTime(value) : t('admin.ipBans.neverHit') }}
            </span>
          </template>

          <template #cell-expires_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">
              {{ value ? formatDateTime(value) : t('admin.ipBans.permanent') }}
            </span>
          </template>

          <template #cell-created_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">
              {{ formatDateTime(value) }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center space-x-1">
              <button
                class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-blue-50 hover:text-blue-600 dark:hover:bg-blue-900/20 dark:hover:text-blue-400"
                :title="t('common.edit')"
                @click="openEditDialog(row)"
              >
                <Icon name="edit" size="sm" />
              </button>
              <button
                class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-amber-50 hover:text-amber-600 dark:hover:bg-amber-900/20 dark:hover:text-amber-400"
                :title="row.status === 'active' ? t('admin.ipBans.disable') : t('admin.ipBans.enable')"
                @click="toggleStatus(row)"
              >
                <Icon :name="row.status === 'active' ? 'ban' : 'check'" size="sm" />
              </button>
              <button
                class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
                :title="t('common.delete')"
                @click="openDeleteDialog(row)"
              >
                <Icon name="trash" size="sm" />
              </button>
            </div>
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <BaseDialog
      :show="showFormDialog"
      :title="editingBan ? t('admin.ipBans.editRule') : t('admin.ipBans.createRule')"
      width="normal"
      @close="closeFormDialog"
    >
      <form id="ip-ban-form" class="space-y-4" @submit.prevent="submitForm">
        <div>
          <label class="input-label">{{ t('admin.ipBans.pattern') }}</label>
          <input
            v-model="form.pattern"
            type="text"
            required
            class="input font-mono"
            :placeholder="t('admin.ipBans.patternPlaceholder')"
          />
          <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
            {{ t('admin.ipBans.patternHelp') }}
          </p>
        </div>
        <div v-if="editingBan">
          <label class="input-label">{{ t('admin.ipBans.status') }}</label>
          <Select v-model="form.status" :options="statusOptions" />
        </div>
        <div>
          <label class="input-label">
            {{ t('admin.ipBans.expiresAt') }}
            <span class="ml-1 text-xs font-normal text-gray-400">({{ t('common.optional') }})</span>
          </label>
          <input v-model="form.expires_at_str" type="datetime-local" class="input" />
        </div>
        <div>
          <label class="input-label">
            {{ t('admin.ipBans.reason') }}
            <span class="ml-1 text-xs font-normal text-gray-400">({{ t('common.optional') }})</span>
          </label>
          <textarea
            v-model="form.reason"
            rows="3"
            class="input"
            :placeholder="t('admin.ipBans.reasonPlaceholder')"
          ></textarea>
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="closeFormDialog">
            {{ t('common.cancel') }}
          </button>
          <button type="submit" form="ip-ban-form" class="btn btn-primary" :disabled="submitting">
            {{ submitting ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.ipBans.deleteRule')"
      :message="t('admin.ipBans.deleteConfirm')"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      danger
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { useAppStore } from '@/stores/app'
import type { IPBan, IPBanStatus } from '@/types'
import { formatCompactNumber, formatDateTime, parseDateTimeLocalInput } from '@/utils/format'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

const bans = ref<IPBan[]>([])
const loading = ref(false)
const submitting = ref(false)
const searchQuery = ref('')
const editingBan = ref<IPBan | null>(null)
const deletingBan = ref<IPBan | null>(null)
const showFormDialog = ref(false)
const showDeleteDialog = ref(false)

const filters = reactive({ status: '' })
const pagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0
})
const sortState = reactive({
  sort_by: 'created_at',
  sort_order: 'desc' as 'asc' | 'desc'
})
const form = reactive({
  pattern: '',
  status: 'active' as IPBanStatus,
  expires_at_str: '',
  reason: ''
})

const filterStatusOptions = computed(() => [
  { value: '', label: t('admin.ipBans.allStatus') },
  { value: 'active', label: t('admin.ipBans.statusActive') },
  { value: 'inactive', label: t('admin.ipBans.statusInactive') }
])

const statusOptions = computed(() => [
  { value: 'active', label: t('admin.ipBans.statusActive') },
  { value: 'inactive', label: t('admin.ipBans.statusInactive') }
])

const columns = computed<Column[]>(() => [
  { key: 'pattern', label: t('admin.ipBans.columns.pattern'), sortable: true },
  { key: 'status', label: t('admin.ipBans.columns.status'), sortable: true },
  { key: 'reason', label: t('admin.ipBans.columns.reason') },
  { key: 'source', label: t('admin.ipBans.columns.source') },
  { key: 'hit_count', label: t('admin.ipBans.columns.hitCount'), sortable: true },
  { key: 'last_hit_at', label: t('admin.ipBans.columns.lastHitAt'), sortable: true },
  { key: 'expires_at', label: t('admin.ipBans.columns.expiresAt'), sortable: true },
  { key: 'created_at', label: t('admin.ipBans.columns.createdAt'), sortable: true },
  { key: 'actions', label: t('admin.ipBans.columns.actions') }
])

let abortController: AbortController | null = null
let searchTimeout: ReturnType<typeof setTimeout>

const loadBans = async () => {
  abortController?.abort()
  const currentController = new AbortController()
  abortController = currentController
  loading.value = true
  try {
    const response = await adminAPI.ipBans.list(
      pagination.page,
      pagination.page_size,
      {
        status: filters.status || undefined,
        search: searchQuery.value || undefined,
        sort_by: sortState.sort_by,
        sort_order: sortState.sort_order
      },
      { signal: currentController.signal }
    )
    if (currentController.signal.aborted || abortController !== currentController) return
    bans.value = response.items
    pagination.total = response.total
  } catch (error: any) {
    if (
      currentController.signal.aborted ||
      abortController !== currentController ||
      error?.name === 'AbortError' ||
      error?.code === 'ERR_CANCELED'
    ) {
      return
    }
    appStore.showError(t('admin.ipBans.failedToLoad'))
  } finally {
    if (abortController === currentController) {
      loading.value = false
      abortController = null
    }
  }
}

const handleSearch = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(reloadFromFirstPage, 300)
}

const reloadFromFirstPage = () => {
  pagination.page = 1
  loadBans()
}

const handlePageChange = (page: number) => {
  pagination.page = page
  loadBans()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.page_size = pageSize
  reloadFromFirstPage()
}

const handleSort = (key: string, order: 'asc' | 'desc') => {
  sortState.sort_by = key
  sortState.sort_order = order
  reloadFromFirstPage()
}

const openCreateDialog = () => {
  editingBan.value = null
  resetForm()
  showFormDialog.value = true
}

const openEditDialog = (ban: IPBan) => {
  editingBan.value = ban
  form.pattern = ban.pattern
  form.status = ban.status
  form.reason = ban.reason || ''
  form.expires_at_str = ban.expires_at ? toDateTimeLocal(ban.expires_at) : ''
  showFormDialog.value = true
}

const closeFormDialog = () => {
  showFormDialog.value = false
  editingBan.value = null
  resetForm()
}

const resetForm = () => {
  form.pattern = ''
  form.status = 'active'
  form.expires_at_str = ''
  form.reason = ''
}

const submitForm = async () => {
  submitting.value = true
  try {
    const expiresAt = parseDateTimeLocalInput(form.expires_at_str)
    if (editingBan.value) {
      await adminAPI.ipBans.update(editingBan.value.id, {
        pattern: form.pattern,
        status: form.status,
        reason: form.reason || '',
        expires_at: expiresAt ?? 0
      })
      appStore.showSuccess(t('admin.ipBans.ruleUpdated'))
    } else {
      await adminAPI.ipBans.create({
        pattern: form.pattern,
        reason: form.reason || undefined,
        expires_at: expiresAt
      })
      appStore.showSuccess(t('admin.ipBans.ruleCreated'))
    }
    closeFormDialog()
    loadBans()
  } catch (error: any) {
    appStore.showError(getErrorMessage(error, editingBan.value ? t('admin.ipBans.failedToUpdate') : t('admin.ipBans.failedToCreate')))
  } finally {
    submitting.value = false
  }
}

const toggleStatus = async (ban: IPBan) => {
  const nextStatus: IPBanStatus = ban.status === 'active' ? 'inactive' : 'active'
  try {
    await adminAPI.ipBans.update(ban.id, { status: nextStatus })
    appStore.showSuccess(nextStatus === 'active' ? t('admin.ipBans.ruleEnabled') : t('admin.ipBans.ruleDisabled'))
    loadBans()
  } catch (error: any) {
    appStore.showError(getErrorMessage(error, t('admin.ipBans.failedToUpdate')))
  }
}

const openDeleteDialog = (ban: IPBan) => {
  deletingBan.value = ban
  showDeleteDialog.value = true
}

const confirmDelete = async () => {
  if (!deletingBan.value) return
  try {
    await adminAPI.ipBans.delete(deletingBan.value.id)
    appStore.showSuccess(t('admin.ipBans.ruleDeleted'))
    showDeleteDialog.value = false
    deletingBan.value = null
    loadBans()
  } catch (error: any) {
    appStore.showError(getErrorMessage(error, t('admin.ipBans.failedToDelete')))
  }
}

const getStatusClass = (ban: IPBan) => {
  if (ban.status !== 'active') return 'badge-gray'
  if (isExpired(ban)) return 'badge-danger'
  return 'badge-success'
}

const getStatusLabel = (ban: IPBan) => {
  return ban.status === 'active' ? t('admin.ipBans.statusActive') : t('admin.ipBans.statusInactive')
}

const isExpired = (ban: IPBan) => {
  return Boolean(ban.expires_at && new Date(ban.expires_at) <= new Date())
}

const toDateTimeLocal = (value: string) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day}T${hours}:${minutes}`
}

const getErrorMessage = (error: any, fallback: string) => {
  return error?.response?.data?.message || error?.response?.data?.detail || fallback
}

onMounted(loadBans)

onUnmounted(() => {
  clearTimeout(searchTimeout)
  abortController?.abort()
})
</script>
