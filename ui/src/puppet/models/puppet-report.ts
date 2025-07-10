import { autoImplement, msToTime } from 'src/helper/functions';
import {
  ApiPuppetEventCount,
  PuppetEventCount,
} from 'src/puppet/models/puppet-event-count';
import moment from 'moment';

export interface ApiPuppetReportMetric {
  category: string;
  name: string;
  value: number;
}

export class PuppetReportMetric extends autoImplement<ApiPuppetReportMetric>() {
  static fromApi(apiItem: ApiPuppetReportMetric): PuppetReportMetric {
    return PuppetReportMetric.fromApi(apiItem);
  }
}

export type PuppetReportMetrics = {
  data: ApiPuppetReportMetric[];
  href: string;
};

export type PuppetReportLogs = {
  data: ApiPuppetReportLog[];
  href: string;
};

export interface ApiPuppetReportLog {
  file: string;
  level: string;
  line: number;
  message: string;
  source: string;
  tags: string[];
  time: Date;
}

export class PuppetReportLog extends autoImplement<ApiPuppetReportLog>() {
  static fromApi(apiItem: ApiPuppetReportLog): PuppetReportLog {
    return new PuppetReportLog(apiItem);
  }

  get color() {
    switch (this.level) {
      case 'info':
        return 'primary';
      case 'notice':
        return 'secondary';
      case 'err':
        return 'negative';
      case 'warning':
        return 'warning';
    }
  }
}

export interface ApiPuppetReport {
  cached_catalog_status: string;
  catalog_uuid: string;
  certname: string;
  code_id: null;
  configuration_version: string;
  corrective_change: null;
  start_time: Date;
  end_time: Date;
  environment: string;
  hash: string;
  job_id: null;
  logs: PuppetReportLogs;
  metrics: PuppetReportMetrics;
  noop: boolean;
  noop_pending: boolean;
  producer: string;
  producer_timestamp: Date;
  puppet_version: string;
  receive_time: Date;
  report_format: number;
  status: string;
  transaction_uuid: string;
  type: string;
}

export class PuppetReport extends autoImplement<ApiPuppetReport>() {
  static fromApi(apiItem: ApiPuppetReport): PuppetReport {
    return new PuppetReport(apiItem);
  }

  public getMetricsValue(category: string, name: string) {
    return this.metrics.data.filter(
      (s) => s.category == category && s.name == name
    )[0].value;
  }

  public getEventCounts(): PuppetEventCount {
    return new PuppetEventCount({
      successes: this.getMetricsValue('events', 'success'),
      failures: this.getMetricsValue('events', 'failure'),
      skips: this.getMetricsValue('resources', 'skipped'),
      noops: 0,
    } as ApiPuppetEventCount);
  }

  get endTimeFormatted() {
    return moment(this.end_time).format('DD. MMM. YYYY - HH:mm:ss');
  }

  get durationInMs() {
    return (
      new Date(this.end_time).valueOf() - new Date(this.start_time).valueOf()
    );
  }

  get durationFormatted() {
    return msToTime(this.durationInMs);
  }

  get logsMapped() {
    return this.logs.data.map((s) => PuppetReportLog.fromApi(s));
  }
}
