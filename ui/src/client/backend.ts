import { api } from 'boot/axios';
import type { AxiosPromise } from 'axios';
import type { ApiVersion, BaseResponse } from 'src/client/models';
import type PqlQuery from 'src/puppet/query-builder';
import type {
  ApiPredefinedView, ApiPredefinedViewResult,
  ApiPuppetQueryPredefined,
  PuppetQueryHistoryEntry,
  PuppetQueryRequest,
  PuppetQueryResult
} from 'src/puppet/models';
import { type ApiPuppetNodeWithEventCount } from 'src/puppet/models/puppet-node';

class Backend {
  getQueryResult<T>(query: PqlQuery) : AxiosPromise<BaseResponse<PuppetQueryResult<T>>> {
    const payload = {
      Query: query.build(),
    } as PuppetQueryRequest;

    return api.post('/api/v1/pdb/query', payload);
  }

  getQueryHistory() : AxiosPromise<BaseResponse<PuppetQueryHistoryEntry[]>>{
    return api.get('/api/v1/pdb/query/history');
  }

  getQueryPredefined() : AxiosPromise<BaseResponse<ApiPuppetQueryPredefined[]>>{
    return api.get('/api/v1/pdb/query/predefined');
  }

  getRawQueryResult<T>(query: string, save?: boolean) : AxiosPromise<BaseResponse<PuppetQueryResult<T>>> {
    const payload = {
      query: query,
      saveInHistory: save,
    }

    return api.post('/api/v1/pdb/query', payload);
  }

  getFactNames() : AxiosPromise<BaseResponse<string[]>> {
    return api.get('/api/v1/pdb/fact-names');
  }

  getViewNodeOverview(environment: string, status?: string[]) : AxiosPromise<BaseResponse<ApiPuppetNodeWithEventCount[]>> {
    let queryParams = ''
    if (status) {
      status.forEach(s => {
        queryParams += `&status=${s}`
      })
    }
    if (queryParams != '') queryParams = `${queryParams}`

    return api.get(`/api/v1/view/node_overview?environment=${environment}${queryParams}`)
  }

  getPredefinedViews() : AxiosPromise<BaseResponse<ApiPredefinedView[]>> {
    return api.get('/api/v1/view/predefined')
  }

  getPredefinedViewsResult(viewName: string) : AxiosPromise<BaseResponse<ApiPredefinedViewResult>> {
    return api.get(`/api/v1/view/predefined/${viewName}`)
  }

  getVersion() : AxiosPromise<BaseResponse<ApiVersion>>{
    return api.get('/api/v1/version')
  }
}

export default new Backend();
