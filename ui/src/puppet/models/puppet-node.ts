import { autoImplement } from 'src/helper/functions';
import { ApiPuppetEventCount, PuppetEventCount } from 'src/puppet/models/puppet-event-count';

export interface ApiPuppetNode {
  cached_catalog_status: string;
  catalog_environment: string;
  catalog_timestamp: Date;
  certname: string;
  deactivated: boolean | null;
  expired: boolean | null;
  facts_environment: string;
  facts_timestamp: Date;
  latest_report_corrective_change: boolean | null;
  latest_report_hash: string;
  latest_report_job_id: string;
  latest_report_noop: boolean;
  latest_report_noop_pending: boolean;
  latest_report_status: string;
  report_environment: string;
  report_timestamp: Date;
}

export class PuppetNode extends autoImplement<ApiPuppetNode>() {
  static fromApi(apiItem: ApiPuppetNode) : PuppetNode {
    return new PuppetNode(apiItem);
  }
}

export interface ApiPuppetNodeWithEventCount extends PuppetNode {
  events: ApiPuppetEventCount;
}

export class PuppetNodeWithEventCount extends autoImplement<ApiPuppetNodeWithEventCount>() {
  static fromApi(apiItem: ApiPuppetNodeWithEventCount) : PuppetNodeWithEventCount {
    return new PuppetNodeWithEventCount(apiItem);
  }

  get eventsMapped() {
    return PuppetEventCount.fromApi(this.events)
  }
}
