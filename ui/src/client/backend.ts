import { api } from 'boot/axios';
import type { AxiosPromise } from 'axios';
import type { ApiMeta, ApiVersion, BaseResponse } from 'src/client/models';
import type PqlQuery from 'src/puppet/query-builder';
import type {
  ApiPredefinedView,
  ApiPredefinedViewResult,
  ApiPuppetQueryPredefined,
  PuppetQueryHistoryEntry,
  PuppetQueryRequest,
  PuppetQueryResult,
} from 'src/puppet/models';
import { type ApiPuppetNodeWithEventCount } from 'src/puppet/models/puppet-node';
import type { CertificateStatusResponse, CertificateState, CertificateStatusQuery } from 'src/puppet/models/certificate-status';

class Backend {
  getQueryResult<T>(query: PqlQuery): AxiosPromise<BaseResponse<PuppetQueryResult<T>>> {
    const payload = {
      Query: query.build(),
    } as PuppetQueryRequest;

    return api.post('/api/v1/pdb/query', payload);
  }

  getQueryHistory(): AxiosPromise<BaseResponse<PuppetQueryHistoryEntry[]>> {
    return api.get('/api/v1/pdb/query/history');
  }

  getQueryPredefined(): AxiosPromise<BaseResponse<ApiPuppetQueryPredefined[]>> {
    return api.get('/api/v1/pdb/query/predefined');
  }

  getRawQueryResult<T>(query: string, save?: boolean): AxiosPromise<BaseResponse<PuppetQueryResult<T>>> {
    const payload = {
      query: query,
      saveInHistory: save,
    }

    return api.post('/api/v1/pdb/query', payload);
  }

  getViewNodeOverview(environment?: string, status?: string[]): AxiosPromise<BaseResponse<ApiPuppetNodeWithEventCount[]>> {
    const queryParams = new URLSearchParams();
    if (environment) {
      queryParams.append("environment", environment);
    }
    if (status) {
      status.forEach(s => {
        queryParams.append("status", s);
      })
    }

    return api.get(`/api/v1/view/node_overview?${queryParams}`)
  }

  getPredefinedViews(): AxiosPromise<BaseResponse<ApiPredefinedView[]>> {
    return api.get('/api/v1/view/predefined')
  }

  getPredefinedViewsResult(viewName: string): AxiosPromise<BaseResponse<ApiPredefinedViewResult>> {
    return api.get(`/api/v1/view/predefined/${viewName}`)
  }

  getPredefinedViewsMeta(viewName: string): AxiosPromise<BaseResponse<ApiPredefinedView>> {
    return api.get(`/api/v1/view/predefined/${viewName}/meta`)
  }

  getMeta(): AxiosPromise<BaseResponse<ApiMeta>> {
    return api.get('/api/v1/meta')
  }

  getVersion(): AxiosPromise<BaseResponse<ApiVersion>> {
    return api.get('/api/v1/version')
  }

  getCertificateStatuses(states?: CertificateState[], filter?: string): AxiosPromise<BaseResponse<CertificateStatusResponse>> {
    return api.post('/api/v1/ca/status', {
      states: states,
      filter: filter,
    } as CertificateStatusQuery);
  }

  signCertificate(name: string): AxiosPromise<BaseResponse<null>> {
    return api.post(`/api/v1/ca/status/${name}/sign`);
  }

  revokeCertificate(name: string): AxiosPromise<BaseResponse<null>> {
    return api.post(`/api/v1/ca/status/${name}/revoke`);
  }

  cleanCertificate(name: string): AxiosPromise<BaseResponse<null>> {
    return api.delete(`/api/v1/ca/status/${name}`);
  }
}

export default new Backend();
