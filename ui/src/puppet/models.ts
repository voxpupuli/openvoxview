/* eslint-disable  @typescript-eslint/no-explicit-any */

import { autoImplement } from 'src/helper/functions';

export interface PuppetEnvironment {
  name: string;
}

export interface ApiPuppetFact {
  certname: string;
  environment: string;
  name: string;
  value: any;
}

export class PuppetFact extends autoImplement<ApiPuppetFact>() {
  static fromApi(apiItem: ApiPuppetFact): PuppetFact {
    return new PuppetFact(apiItem);
  }

  get isJson() {
    return typeof this.value === 'object' && this.value !== null;
  }
}

export type PuppetQueryRequest = {
  Query: string;
};

export interface PuppetQueryResult<T> {
  Data: T;
  Error: string;
  Success: boolean;
  ExecutedOn: Date;
  ExecutionTimeInMilli: number;
}

export type PuppetQueryHistoryEntry = {
  Query: PuppetQueryRequest;
  Result: PuppetQueryResult<unknown[]>;
};

export interface ApiPuppetQueryPredefined {
  Description: string;
  Query: string;
}

export class PuppetQueryPredefined extends autoImplement<ApiPuppetQueryPredefined>() {
  static fromApi(apiItem: ApiPuppetQueryPredefined): PuppetQueryPredefined {
    return new PuppetQueryPredefined(apiItem);
  }
}

export interface ApiPredefinedViewResult {
  View: ApiPredefinedView;
  Data: unknown[];
}

export class PredefinedViewResult extends autoImplement<ApiPredefinedViewResult>() {
  static fromApi(apiItem: ApiPredefinedViewResult): PredefinedViewResult {
    return new PredefinedViewResult(apiItem);
  }
}

export interface ApiPredefinedViewFact {
  Name: string;
  Fact: string;
  Renderer: string;
}

export interface ApiPredefinedView {
  Name: string;
  Facts: ApiPredefinedViewFact[];
  RowsPerPage: number;
}

export class PredefinedView extends autoImplement<ApiPredefinedView>() {
  static fromApi(apiItem: ApiPredefinedView): PredefinedView {
    return new PredefinedView(apiItem);
  }
}
