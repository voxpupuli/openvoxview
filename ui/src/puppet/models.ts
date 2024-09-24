/* eslint-disable  @typescript-eslint/no-explicit-any */

import { autoImplement } from 'src/helper/functions';

export interface PuppetEnvironment {
  name: string;
}

export type PuppetFact = {
  certname: string;
  environment: string;
  name: string;
  value: any;
};

export type PuppetQueryRequest = {
  Query: string;
};

export type PuppetQueryResult<T> = {
  Data: T;
  Error: string;
  Success: boolean;
  ExecutedOn: Date;
  ExecutionTimeInMilli: number;
};

export type PuppetQueryHistoryEntry = {
  Query: PuppetQueryRequest;
  Result: PuppetQueryResult<unknown[]>;
};

export interface ApiPuppetQueryPredefined {
  Description: string;
  Query: string;
}

export class PuppetQueryPredefined extends autoImplement<ApiPuppetQueryPredefined>() {
  static fromApi(apiItem: ApiPuppetQueryPredefined) : PuppetQueryPredefined {
    return new PuppetQueryPredefined(apiItem);
  }
}

export interface ApiPredefinedViewResult {
  View: ApiPredefinedView;
  Data: unknown[];
}

export class PredefinedViewResult extends autoImplement<ApiPredefinedViewResult>() {
  static fromApi(apiItem: ApiPredefinedViewResult) : PredefinedViewResult {
    return new PredefinedViewResult(apiItem);
  }
}

export interface ApiPredefinedViewFact {
  Name: string;
  Fact: string;
  Renderer: string;
}

export interface ApiPredefinedView  {
  Name: string;
  Facts: ApiPredefinedViewFact[];
}

export class PredefinedView extends autoImplement<ApiPredefinedView>() {
  static fromApi(apiItem: ApiPredefinedView) : PredefinedView {
    return new PredefinedView(apiItem)
  }
}
