<script setup lang="ts">
import { copyToClipboard, Notify, type QTableColumn, type QTableProps } from 'quasar';
import { CertificateState, type CertificateStatus, CertificateStatusResponse } from 'src/puppet/models/certificate-status';
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import Backend from 'src/client/backend';
import NodeLink from 'components/NodeLink.vue';
import type { ApiMeta } from 'src/client/models';

const DEBOUNCE = 300;

const { t } = useI18n();
const meta = ref<ApiMeta>();
const certificates = ref<CertificateStatus[]>([]);
const filterExpanded = ref(true);
const filter = ref('');

const filterStates = ref(['requested', 'signed']);
const filterOptions = ref([
  {
    label: t('LABEL_REQUESTED'),
    value: 'requested',
  },
  {
    label: t('LABEL_SIGNED'),
    value: 'signed',
  },
  {
    label: t('LABEL_REVOKED'),
    value: 'revoked',
  },
]);

const pagination = ref<NonNullable<QTableProps['pagination']>>({
  sortBy: 'name',
  rowsPerPage: 20,
});

const columnsReadOnly: QTableColumn[] = [
  {
    name: 'name',
    field: 'name',
    label: t('LABEL_NAME'),
    align: 'left',
    sortable: true,
  },
  {
    name: 'alt_names',
    field: 'dns_alt_names',
    label: t('LABEL_ALT_NAMES'),
    align: 'left',
    sortable: true,
    format: (val: CertificateStatus["dns_alt_names"]) => val.map((n) => n.replace(/^DNS:/, "")).join(", "),
  },
  {
    name: 'fingerprint',
    field: 'fingerprint',
    label: t('LABEL_FINGERPRINT'),
    align: 'left',
    sortable: true,
  },
  {
    name: 'expires',
    field: 'not_after',
    label: t('LABEL_EXPIRES'),
    align: 'left',
    sortable: true,
    format: (val: CertificateStatus["not_after"]) => val ? new Date(val).toLocaleString(undefined, {}) : "—",
  },
  {
    name: 'state',
    field: 'state',
    label: t('LABEL_STATE'),
    align: 'left',
    sortable: true,
  },
];

const columnsReadWrite: QTableColumn[] = [
  ...columnsReadOnly,
  {
    name: 'actions',
    field: 'actions',
    label: '',
    align: 'left',
    sortable: false,
    classes: 'no-wrap',
  },
];

const columns = computed((): QTableColumn[] => { return meta.value?.CaReadOnly ? columnsReadOnly : columnsReadWrite });

function loadCertificates() {
  const states = filterStates.value.length > 0 ? filterStates.value.map(value => {
    if (!Object.values(CertificateState).includes(value as CertificateState)) {
      throw new Error(`Invalid CertificateState: ${value}`);
    }
    return value as CertificateState;
  }) : undefined;

  const filterBy = filter.value.trim() !== '' ? filter.value.trim() : undefined;

  void Backend.getCertificateStatuses(states, filterBy).then((result) => {
    if (result.status === 200) {
      const resp = CertificateStatusResponse.fromApi(result.data.Data);
      certificates.value = resp.certificate_statuses || [];
    }
  });
}

function loadMeta() {
  void Backend.getMeta().then((result) => {
    if (result.status === 200) {
      meta.value = result.data.Data;
    }
  });
}

function signCertificate(name: string) {
  void Backend.signCertificate(name).then((result) => {
    if (result.status === 200) {
      setTimeout(() => loadCertificates(), DEBOUNCE);
    }
  });
}

function revokeCertificate(name: string) {
  void Backend.revokeCertificate(name).then((result) => {
    if (result.status === 200) {
      setTimeout(() => loadCertificates(), DEBOUNCE);
    }
  });
}

function cleanCertificate(name: string) {
  void Backend.cleanCertificate(name).then((result) => {
    if (result.status === 200) {
      setTimeout(() => loadCertificates(), DEBOUNCE);
    }
  });
}

function copy(text: string) {
  copyToClipboard(text)
    .then(() => {
      Notify.create({
        message: 'Copied',
        color: 'positive',
        position: 'top'
      })
    })
    .catch(() => {
      Notify.create({
        message: 'Copy failed',
        color: 'negative',
        position: 'top'
      })
    })
}

watch(filterStates, () => {
  loadCertificates();
});

onMounted(() => {
  loadMeta();
  loadCertificates();
});
</script>

<style scoped>
.col-fingerprint {
  font-family: monospace;
  font-size: 90%;
  max-width: 15em;
}
</style>
<style>
.tt-ca-fpr {
  font-family: monospace;
  font-size: 100%;
}
</style>

<template>
  <q-page padding>
    <q-card>
      <q-card-section class="bg-primary text-white text-h6">
        <div class="row">
          {{ $t('LABEL_FILTER') }}
          <q-space />
          <q-btn color="grey" round flat dense :icon="filterExpanded ? 'keyboard_arrow_up' : 'keyboard_arrow_down'"
            @click="filterExpanded = !filterExpanded" />
        </div>
      </q-card-section>
      <q-slide-transition>
        <div v-show="filterExpanded">
          <q-card-section>
            <q-input :debounce="DEBOUNCE" v-model="filter" :placeholder="$t('LABEL_SEARCH')" class="full-width"
              @update:model-value="loadCertificates()" />
            <q-select :label="$t('LABEL_STATE')" v-model="filterStates" :options="filterOptions" multiple use-chips
              map-options emit-value class="full-width">
              <template v-slot:option="{ itemProps, opt, selected, toggleOption }">
                <q-item v-bind="itemProps">
                  <q-item-section>
                    <q-item-label>{{ opt.label }}</q-item-label>
                  </q-item-section>
                  <q-item-section side>
                    <q-toggle :model-value="selected" @update:model-value="toggleOption(opt)" />
                  </q-item-section>
                </q-item>
              </template>
            </q-select>
          </q-card-section>
        </div>
      </q-slide-transition>
    </q-card>
    <q-table :rows="certificates" :columns="columns" table-header-class="bg-primary text-white" :pagination="pagination"
      wrap-cells class="q-mt-md" :title="$t('LABEL_CERTIFICATES', 2)">
      <template v-slot:top-right>
        <q-btn icon="refresh" color="secondary" @click="loadCertificates" />
      </template>
      <template v-slot:body="props">
        <q-tr :props="props">
          <q-td v-for="col in props.cols" :key="col.name" :props="props">
            <div v-if="col.name == 'name'">
              <NodeLink :certname="col.value" />
            </div>
            <div v-else-if="col.name == 'fingerprint'">
              <div class="ellipsis col-fingerprint">
                <q-btn icon="content_copy" flat dense size="sm" @click="copy(col.value)" />
                {{ col.value }}
                <q-tooltip class="bg-dark text-white tt-ca-fpr">{{ col.value }}</q-tooltip>
              </div>
            </div>
            <div v-else-if="col.name == 'actions'" class="row no-wrap items-center q-gutter-xs">
              <div v-if="props.row.state === 'requested'">
                <q-btn dense color="positive" icon="check_circle" @click="signCertificate(props.row.name)">
                  <q-tooltip>
                    Sign
                  </q-tooltip>
                </q-btn>
              </div>
              <div v-if="props.row.state === 'signed'">
                <q-btn dense color="primary" icon="cancel" @click="revokeCertificate(props.row.name)">
                  <q-tooltip>
                    Revoke
                  </q-tooltip>
                </q-btn>
              </div>
              <q-btn dense color="negative" icon="remove_circle" @click="cleanCertificate(props.row.name)">
                <q-tooltip>
                  Clean
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
  </q-page>
</template>
