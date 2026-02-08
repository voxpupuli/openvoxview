import { autoImplement } from 'src/helper/functions';
import { type ApiPuppetEventCount, PuppetEventCount } from 'src/puppet/models/puppet-event-count';

export interface ApiPuppetNode {
  // Fields in OpenVoxDB nodes response format:
  // https://github.com/OpenVoxProject/openvoxdb/blob/8.12.1/documentation/api/query/v4/nodes.markdown#response-format
  certname: string;
  deactivated: Date | null;
  expired: Date | null;
  catalog_timestamp: Date | null;
  facts_timestamp: Date | null;
  report_timestamp: Date | null;
  catalog_environment: string | null;
  facts_environment: string | null;
  report_environment: string | null;
  latest_report_status: string;
  latest_report_noop: boolean;
  latest_report_noop_pending: boolean;
  latest_report_hash: string;
  latest_report_job_id: string;

  // Fields not in the docs linked above, but are in the OpenVoxDB API response:
  cached_catalog_status: string | null;
  latest_report_corrective_change: boolean | null;
}

export class PuppetNode extends autoImplement<ApiPuppetNode>() {
  static fromApi(apiItem: ApiPuppetNode): PuppetNode {
    return new PuppetNode(apiItem);
  }
}

export interface ApiPuppetNodeWithEventCount extends PuppetNode {
  events: ApiPuppetEventCount;
}

export class PuppetNodeWithEventCount extends autoImplement<ApiPuppetNodeWithEventCount>() {
  static fromApi(apiItem: ApiPuppetNodeWithEventCount): PuppetNodeWithEventCount {
    return new PuppetNodeWithEventCount(apiItem);
  }

  get eventsMapped() {
    return PuppetEventCount.fromApi(this.events)
  }
}
