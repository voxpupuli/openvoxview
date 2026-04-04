<script setup lang="ts">
import { onMounted, ref, computed } from 'vue';
import { useQuasar, type QTableColumn, type QTableProps, Notify } from 'quasar';
import { useI18n } from 'vue-i18n';
import Backend from 'src/client/backend';
import type { UserProfile } from 'src/client/models';
import { useAuthStore } from 'stores/auth';

const q = useQuasar();
const { t } = useI18n();
const auth = useAuthStore();

const users = ref<UserProfile[]>([]);
const showDialog = ref(false);
const editingUser = ref<UserProfile | null>(null);

const formUsername = ref('');
const formEmail = ref('');
const formDisplayName = ref('');
const formPassword = ref('');
const formPasswordConfirm = ref('');
const formError = ref('');
const formLoading = ref(false);

const currentUserId = computed(() => {
  if (!auth.accessToken) return null;
  try {
    const payload = JSON.parse(atob(auth.accessToken.split('.')[1] ?? ''));
    return Number(payload.sub);
  } catch {
    return null;
  }
});

const isEditing = computed(() => editingUser.value !== null);
const isSamlUser = computed(() => editingUser.value?.auth_source === 'saml');

const dialogTitle = computed(() =>
  isEditing.value ? t('LABEL_EDIT_USER') : t('LABEL_CREATE_USER'),
);

const pagination = ref<NonNullable<QTableProps['pagination']>>({
  sortBy: 'username',
  rowsPerPage: 20,
});

const columns: QTableColumn[] = [
  {
    name: 'username',
    field: 'username',
    label: t('LABEL_USERNAME'),
    align: 'left',
    sortable: true,
  },
  {
    name: 'display_name',
    field: 'display_name',
    label: t('LABEL_DISPLAY_NAME'),
    align: 'left',
    sortable: true,
  },
  {
    name: 'email',
    field: 'email',
    label: t('LABEL_EMAIL'),
    align: 'left',
    sortable: true,
  },
  {
    name: 'auth_source',
    field: 'auth_source',
    label: t('LABEL_AUTH_SOURCE'),
    align: 'left',
    sortable: true,
  },
  {
    name: 'created_at',
    field: 'created_at',
    label: t('LABEL_CREATED'),
    align: 'left',
    sortable: true,
    format: (val: string) => new Date(val).toLocaleString(),
  },
  {
    name: 'actions',
    field: 'actions',
    label: '',
    align: 'right',
    sortable: false,
  },
];

function loadUsers() {
  void Backend.getUsers().then((result) => {
    if (result.status === 200) {
      users.value = result.data.Data;
    }
  });
}

function openCreateDialog() {
  editingUser.value = null;
  formUsername.value = '';
  formEmail.value = '';
  formDisplayName.value = '';
  formPassword.value = '';
  formPasswordConfirm.value = '';
  formError.value = '';
  showDialog.value = true;
}

function openEditDialog(user: UserProfile) {
  editingUser.value = user;
  formUsername.value = user.username;
  formEmail.value = user.email || '';
  formDisplayName.value = user.display_name || '';
  formPassword.value = '';
  formPasswordConfirm.value = '';
  formError.value = '';
  showDialog.value = true;
}

async function onSubmit() {
  formError.value = '';

  if (!isEditing.value && !formUsername.value.trim()) {
    formError.value = t('LABEL_USERNAME_REQUIRED');
    return;
  }

  if (!isEditing.value && formPassword.value.length < 8) {
    formError.value = t('LABEL_PASSWORD_MIN_LENGTH');
    return;
  }

  if (formPassword.value && formPassword.value !== formPasswordConfirm.value) {
    formError.value = t('LABEL_PASSWORDS_NO_MATCH');
    return;
  }

  if (isEditing.value && formPassword.value && formPassword.value.length < 8) {
    formError.value = t('LABEL_PASSWORD_MIN_LENGTH');
    return;
  }

  formLoading.value = true;

  try {
    if (isEditing.value && editingUser.value) {
      const data: { email?: string; display_name?: string; password?: string } = {
        email: formEmail.value,
        display_name: formDisplayName.value,
      };
      if (formPassword.value) {
        data.password = formPassword.value;
      }
      await Backend.updateUser(editingUser.value.id, data);
      Notify.create({ message: t('NOTIFICATION_USER_UPDATED'), color: 'positive' });
    } else {
      const createData: { username: string; password: string; email?: string; display_name?: string } = {
        username: formUsername.value.trim(),
        password: formPassword.value,
      };
      if (formEmail.value) createData.email = formEmail.value;
      if (formDisplayName.value) createData.display_name = formDisplayName.value;
      await Backend.createUser(createData);
      Notify.create({ message: t('NOTIFICATION_USER_CREATED'), color: 'positive' });
    }
    showDialog.value = false;
    loadUsers();
  } catch (error: unknown) {
    const axiosError = error as { response?: { status?: number; data?: { Error?: string } } };
    if (axiosError.response?.status === 409) {
      formError.value = t('LABEL_USERNAME_EXISTS');
    } else if (axiosError.response?.data?.Error) {
      formError.value = axiosError.response.data.Error;
    } else {
      formError.value = t('LABEL_ERROR');
    }
  } finally {
    formLoading.value = false;
  }
}

function confirmDelete(user: UserProfile) {
  q.dialog({
    title: t('LABEL_CONFIRM_DELETE_USER_TITLE'),
    message: t('LABEL_CONFIRM_DELETE_USER_MESSAGE', { username: user.username }),
    cancel: true,
    ok: {
      label: t('LABEL_DELETE'),
      color: 'negative',
    },
  }).onOk(() => {
    void Backend.deleteUser(user.id).then(() => {
      Notify.create({ message: t('NOTIFICATION_USER_DELETED'), color: 'positive' });
      loadUsers();
    });
  });
}

onMounted(() => {
  loadUsers();
});
</script>

<template>
  <q-page padding>
    <q-table
      :rows="users"
      :columns="columns"
      row-key="id"
      table-header-class="bg-primary text-white"
      :pagination="pagination"
      :title="$t('MENU_USERS')"
    >
      <template v-slot:top-right>
        <q-btn icon="person_add" color="primary" :label="$t('BTN_ADD_USER')" @click="openCreateDialog" />
        <q-btn icon="refresh" color="secondary" class="q-ml-sm" @click="loadUsers" />
      </template>
      <template v-slot:body="props">
        <q-tr :props="props">
          <q-td v-for="col in props.cols" :key="col.name" :props="props">
            <div v-if="col.name === 'actions'" class="row no-wrap items-center justify-end q-gutter-xs">
              <q-btn dense flat icon="edit" color="primary" @click="openEditDialog(props.row)">
                <q-tooltip>{{ $t('BTN_EDIT') }}</q-tooltip>
              </q-btn>
              <q-btn
                dense
                flat
                icon="delete"
                color="negative"
                :disable="props.row.id === currentUserId"
                @click="confirmDelete(props.row)"
              >
                <q-tooltip>
                  {{ props.row.id === currentUserId ? $t('LABEL_CANNOT_DELETE_SELF') : $t('LABEL_DELETE') }}
                </q-tooltip>
              </q-btn>
            </div>
            <div v-else>
              {{ col.value }}
            </div>
          </q-td>
        </q-tr>
      </template>
    </q-table>

    <q-dialog v-model="showDialog" persistent>
      <q-card style="min-width: 400px">
        <q-card-section>
          <div class="text-h6">{{ dialogTitle }}</div>
        </q-card-section>

        <q-card-section>
          <q-form @submit="onSubmit" class="q-gutter-sm">
            <q-banner v-if="isSamlUser" dense rounded class="bg-info text-white q-mb-sm">
              <template v-slot:avatar>
                <q-icon name="info" />
              </template>
              {{ $t('LABEL_SAML_USER_MANAGED_BY_IDP') }}
            </q-banner>

            <q-input
              v-model="formUsername"
              :label="$t('LABEL_USERNAME')"
              outlined
              dense
              :disable="isEditing"
              :readonly="isEditing"
            />

            <q-input
              v-model="formEmail"
              :label="$t('LABEL_EMAIL')"
              type="email"
              outlined
              dense
              :readonly="isSamlUser"
              :disable="isSamlUser"
            />

            <q-input
              v-model="formDisplayName"
              :label="$t('LABEL_DISPLAY_NAME')"
              outlined
              dense
              :readonly="isSamlUser"
              :disable="isSamlUser"
            />

            <template v-if="!isSamlUser">
              <q-input
                v-model="formPassword"
                :label="isEditing ? $t('LABEL_NEW_PASSWORD') : $t('LABEL_PASSWORD')"
                type="password"
                outlined
                dense
                :hint="isEditing ? $t('LABEL_PASSWORD_HINT_EDIT') : ''"
              />

              <q-input
                v-model="formPasswordConfirm"
                :label="$t('LABEL_CONFIRM_PASSWORD')"
                type="password"
                outlined
                dense
                v-if="formPassword"
              />
            </template>

            <q-banner v-if="formError" dense rounded class="bg-negative text-white q-mt-sm">
              {{ formError }}
            </q-banner>

            <div class="row justify-end q-gutter-sm q-mt-md">
              <q-btn flat :label="$t('BTN_CLOSE')" v-close-popup />
              <q-btn
                v-if="!isSamlUser"
                type="submit"
                color="primary"
                :label="isEditing ? $t('BTN_SAVE') : $t('BTN_CREATE')"
                :loading="formLoading"
              />
            </div>
          </q-form>
        </q-card-section>
      </q-card>
    </q-dialog>
  </q-page>
</template>
